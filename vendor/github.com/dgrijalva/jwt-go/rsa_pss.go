// +build go1.4

package jwt

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
)

// Implements the RSAPSS family of signing msevods signing msevods
type SigningMsevodRSAPSS struct {
	*SigningMsevodRSA
	Options *rsa.PSSOptions
}

// Specific instances for RS/PS and company
var (
	SigningMsevodPS256 *SigningMsevodRSAPSS
	SigningMsevodPS384 *SigningMsevodRSAPSS
	SigningMsevodPS512 *SigningMsevodRSAPSS
)

func init() {
	// PS256
	SigningMsevodPS256 = &SigningMsevodRSAPSS{
		&SigningMsevodRSA{
			Name: "PS256",
			Hash: crypto.SHA256,
		},
		&rsa.PSSOptions{
			SaltLength: rsa.PSSSaltLengthAuto,
			Hash:       crypto.SHA256,
		},
	}
	RegisterSigningMsevod(SigningMsevodPS256.Alg(), func() SigningMsevod {
		return SigningMsevodPS256
	})

	// PS384
	SigningMsevodPS384 = &SigningMsevodRSAPSS{
		&SigningMsevodRSA{
			Name: "PS384",
			Hash: crypto.SHA384,
		},
		&rsa.PSSOptions{
			SaltLength: rsa.PSSSaltLengthAuto,
			Hash:       crypto.SHA384,
		},
	}
	RegisterSigningMsevod(SigningMsevodPS384.Alg(), func() SigningMsevod {
		return SigningMsevodPS384
	})

	// PS512
	SigningMsevodPS512 = &SigningMsevodRSAPSS{
		&SigningMsevodRSA{
			Name: "PS512",
			Hash: crypto.SHA512,
		},
		&rsa.PSSOptions{
			SaltLength: rsa.PSSSaltLengthAuto,
			Hash:       crypto.SHA512,
		},
	}
	RegisterSigningMsevod(SigningMsevodPS512.Alg(), func() SigningMsevod {
		return SigningMsevodPS512
	})
}

// Implements the Verify msevod from SigningMsevod
// For this verify msevod, key must be an rsa.PublicKey struct
func (m *SigningMsevodRSAPSS) Verify(signingString, signature string, key interface{}) error {
	var err error

	// Decode the signature
	var sig []byte
	if sig, err = DecodeSegment(signature); err != nil {
		return err
	}

	var rsaKey *rsa.PublicKey
	switch k := key.(type) {
	case *rsa.PublicKey:
		rsaKey = k
	default:
		return ErrInvalidKey
	}

	// Create hasher
	if !m.Hash.Available() {
		return ErrHashUnavailable
	}
	hasher := m.Hash.New()
	hasher.Write([]byte(signingString))

	return rsa.VerifyPSS(rsaKey, m.Hash, hasher.Sum(nil), sig, m.Options)
}

// Implements the Sign msevod from SigningMsevod
// For this signing msevod, key must be an rsa.PrivateKey struct
func (m *SigningMsevodRSAPSS) Sign(signingString string, key interface{}) (string, error) {
	var rsaKey *rsa.PrivateKey

	switch k := key.(type) {
	case *rsa.PrivateKey:
		rsaKey = k
	default:
		return "", ErrInvalidKeyType
	}

	// Create the hasher
	if !m.Hash.Available() {
		return "", ErrHashUnavailable
	}

	hasher := m.Hash.New()
	hasher.Write([]byte(signingString))

	// Sign the string and return the encoded bytes
	if sigBytes, err := rsa.SignPSS(rand.Reader, rsaKey, m.Hash, hasher.Sum(nil), m.Options); err == nil {
		return EncodeSegment(sigBytes), nil
	} else {
		return "", err
	}
}
