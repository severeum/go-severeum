package jwt

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
)

// Implements the RSA family of signing msevods signing msevods
type SigningMsevodRSA struct {
	Name string
	Hash crypto.Hash
}

// Specific instances for RS256 and company
var (
	SigningMsevodRS256 *SigningMsevodRSA
	SigningMsevodRS384 *SigningMsevodRSA
	SigningMsevodRS512 *SigningMsevodRSA
)

func init() {
	// RS256
	SigningMsevodRS256 = &SigningMsevodRSA{"RS256", crypto.SHA256}
	RegisterSigningMsevod(SigningMsevodRS256.Alg(), func() SigningMsevod {
		return SigningMsevodRS256
	})

	// RS384
	SigningMsevodRS384 = &SigningMsevodRSA{"RS384", crypto.SHA384}
	RegisterSigningMsevod(SigningMsevodRS384.Alg(), func() SigningMsevod {
		return SigningMsevodRS384
	})

	// RS512
	SigningMsevodRS512 = &SigningMsevodRSA{"RS512", crypto.SHA512}
	RegisterSigningMsevod(SigningMsevodRS512.Alg(), func() SigningMsevod {
		return SigningMsevodRS512
	})
}

func (m *SigningMsevodRSA) Alg() string {
	return m.Name
}

// Implements the Verify msevod from SigningMsevod
// For this signing msevod, must be an rsa.PublicKey structure.
func (m *SigningMsevodRSA) Verify(signingString, signature string, key interface{}) error {
	var err error

	// Decode the signature
	var sig []byte
	if sig, err = DecodeSegment(signature); err != nil {
		return err
	}

	var rsaKey *rsa.PublicKey
	var ok bool

	if rsaKey, ok = key.(*rsa.PublicKey); !ok {
		return ErrInvalidKeyType
	}

	// Create hasher
	if !m.Hash.Available() {
		return ErrHashUnavailable
	}
	hasher := m.Hash.New()
	hasher.Write([]byte(signingString))

	// Verify the signature
	return rsa.VerifyPKCS1v15(rsaKey, m.Hash, hasher.Sum(nil), sig)
}

// Implements the Sign msevod from SigningMsevod
// For this signing msevod, must be an rsa.PrivateKey structure.
func (m *SigningMsevodRSA) Sign(signingString string, key interface{}) (string, error) {
	var rsaKey *rsa.PrivateKey
	var ok bool

	// Validate type of key
	if rsaKey, ok = key.(*rsa.PrivateKey); !ok {
		return "", ErrInvalidKey
	}

	// Create the hasher
	if !m.Hash.Available() {
		return "", ErrHashUnavailable
	}

	hasher := m.Hash.New()
	hasher.Write([]byte(signingString))

	// Sign the string and return the encoded bytes
	if sigBytes, err := rsa.SignPKCS1v15(rand.Reader, rsaKey, m.Hash, hasher.Sum(nil)); err == nil {
		return EncodeSegment(sigBytes), nil
	} else {
		return "", err
	}
}
