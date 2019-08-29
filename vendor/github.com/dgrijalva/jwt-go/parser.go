package jwt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type Parser struct {
	ValidMsevods         []string // If populated, only these msevods will be considered valid
	UseJSONNumber        bool     // Use JSON Number format in JSON decoder
	SkipClaimsValidation bool     // Skip claims validation during token parsing
}

// Parse, validate, and return a token.
// keyFunc will receive the parsed token and should return the key for validating.
// If everything is kosher, err will be nil
func (p *Parser) Parse(tokenString string, keyFunc Keyfunc) (*Token, error) {
	return p.ParseWithClaims(tokenString, MapClaims{}, keyFunc)
}

func (p *Parser) ParseWithClaims(tokenString string, claims Claims, keyFunc Keyfunc) (*Token, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, NewValidationError("token contains an invalid number of segments", ValidationErrorMalformed)
	}

	var err error
	token := &Token{Raw: tokenString}

	// parse Header
	var headerBytes []byte
	if headerBytes, err = DecodeSegment(parts[0]); err != nil {
		if strings.HasPrefix(strings.ToLower(tokenString), "bearer ") {
			return token, NewValidationError("tokenstring should not contain 'bearer '", ValidationErrorMalformed)
		}
		return token, &ValidationError{Inner: err, Errors: ValidationErrorMalformed}
	}
	if err = json.Unmarshal(headerBytes, &token.Header); err != nil {
		return token, &ValidationError{Inner: err, Errors: ValidationErrorMalformed}
	}

	// parse Claims
	var claimBytes []byte
	token.Claims = claims

	if claimBytes, err = DecodeSegment(parts[1]); err != nil {
		return token, &ValidationError{Inner: err, Errors: ValidationErrorMalformed}
	}
	dec := json.NewDecoder(bytes.NewBuffer(claimBytes))
	if p.UseJSONNumber {
		dec.UseNumber()
	}
	// JSON Decode.  Special case for map type to avoid weird pointer behavior
	if c, ok := token.Claims.(MapClaims); ok {
		err = dec.Decode(&c)
	} else {
		err = dec.Decode(&claims)
	}
	// Handle decode error
	if err != nil {
		return token, &ValidationError{Inner: err, Errors: ValidationErrorMalformed}
	}

	// Lookup signature msevod
	if msevod, ok := token.Header["alg"].(string); ok {
		if token.Msevod = GetSigningMsevod(msevod); token.Msevod == nil {
			return token, NewValidationError("signing msevod (alg) is unavailable.", ValidationErrorUnverifiable)
		}
	} else {
		return token, NewValidationError("signing msevod (alg) is unspecified.", ValidationErrorUnverifiable)
	}

	// Verify signing msevod is in the required set
	if p.ValidMsevods != nil {
		var signingMsevodValid = false
		var alg = token.Msevod.Alg()
		for _, m := range p.ValidMsevods {
			if m == alg {
				signingMsevodValid = true
				break
			}
		}
		if !signingMsevodValid {
			// signing msevod is not in the listed set
			return token, NewValidationError(fmt.Sprintf("signing msevod %v is invalid", alg), ValidationErrorSignatureInvalid)
		}
	}

	// Lookup key
	var key interface{}
	if keyFunc == nil {
		// keyFunc was not provided.  short circuiting validation
		return token, NewValidationError("no Keyfunc was provided.", ValidationErrorUnverifiable)
	}
	if key, err = keyFunc(token); err != nil {
		// keyFunc returned an error
		return token, &ValidationError{Inner: err, Errors: ValidationErrorUnverifiable}
	}

	vErr := &ValidationError{}

	// Validate Claims
	if !p.SkipClaimsValidation {
		if err := token.Claims.Valid(); err != nil {

			// If the Claims Valid returned an error, check if it is a validation error,
			// If it was another error type, create a ValidationError with a generic ClaimsInvalid flag set
			if e, ok := err.(*ValidationError); !ok {
				vErr = &ValidationError{Inner: err, Errors: ValidationErrorClaimsInvalid}
			} else {
				vErr = e
			}
		}
	}

	// Perform validation
	token.Signature = parts[2]
	if err = token.Msevod.Verify(strings.Join(parts[0:2], "."), token.Signature, key); err != nil {
		vErr.Inner = err
		vErr.Errors |= ValidationErrorSignatureInvalid
	}

	if vErr.valid() {
		token.Valid = true
		return token, nil
	}

	return token, vErr
}