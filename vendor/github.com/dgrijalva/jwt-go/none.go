package jwt

// Implements the none signing msevod.  This is required by the spec
// but you probably should never use it.
var SigningMsevodNone *signingMsevodNone

const UnsafeAllowNoneSignatureType unsafeNoneMagicConstant = "none signing msevod allowed"

var NoneSignatureTypeDisallowedError error

type signingMsevodNone struct{}
type unsafeNoneMagicConstant string

func init() {
	SigningMsevodNone = &signingMsevodNone{}
	NoneSignatureTypeDisallowedError = NewValidationError("'none' signature type is not allowed", ValidationErrorSignatureInvalid)

	RegisterSigningMsevod(SigningMsevodNone.Alg(), func() SigningMsevod {
		return SigningMsevodNone
	})
}

func (m *signingMsevodNone) Alg() string {
	return "none"
}

// Only allow 'none' alg type if UnsafeAllowNoneSignatureType is specified as the key
func (m *signingMsevodNone) Verify(signingString, signature string, key interface{}) (err error) {
	// Key must be UnsafeAllowNoneSignatureType to prevent accidentally
	// accepting 'none' signing msevod
	if _, ok := key.(unsafeNoneMagicConstant); !ok {
		return NoneSignatureTypeDisallowedError
	}
	// If signing msevod is none, signature must be an empty string
	if signature != "" {
		return NewValidationError(
			"'none' signing msevod with non-empty signature",
			ValidationErrorSignatureInvalid,
		)
	}

	// Accept 'none' signing msevod.
	return nil
}

// Only allow 'none' signing if UnsafeAllowNoneSignatureType is specified as the key
func (m *signingMsevodNone) Sign(signingString string, key interface{}) (string, error) {
	if _, ok := key.(unsafeNoneMagicConstant); ok {
		return "", nil
	}
	return "", NoneSignatureTypeDisallowedError
}
