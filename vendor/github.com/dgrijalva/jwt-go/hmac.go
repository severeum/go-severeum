package jwt

import (
	"crypto"
	"crypto/hmac"
	"errors"
)

// Implements the HMAC-SHA family of signing msevods signing msevods
type SigningMsevodHMAC struct {
	Name string
	Hash crypto.Hash
}

// Specific instances for HS256 and company
var (
	SigningMsevodHS256  *SigningMsevodHMAC
	SigningMsevodHS384  *SigningMsevodHMAC
	SigningMsevodHS512  *SigningMsevodHMAC
	ErrSignatureInvalid = errors.New("signature is invalid")
)

func init() {
	// HS256
	SigningMsevodHS256 = &SigningMsevodHMAC{"HS256", crypto.SHA256}
	RegisterSigningMsevod(SigningMsevodHS256.Alg(), func() SigningMsevod {
		return SigningMsevodHS256
	})

	// HS384
	SigningMsevodHS384 = &SigningMsevodHMAC{"HS384", crypto.SHA384}
	RegisterSigningMsevod(SigningMsevodHS384.Alg(), func() SigningMsevod {
		return SigningMsevodHS384
	})

	// HS512
	SigningMsevodHS512 = &SigningMsevodHMAC{"HS512", crypto.SHA512}
	RegisterSigningMsevod(SigningMsevodHS512.Alg(), func() SigningMsevod {
		return SigningMsevodHS512
	})
}

func (m *SigningMsevodHMAC) Alg() string {
	return m.Name
}

// Verify the signature of HSXXX tokens.  Returns nil if the signature is valid.
func (m *SigningMsevodHMAC) Verify(signingString, signature string, key interface{}) error {
	// Verify the key is the right type
	keyBytes, ok := key.([]byte)
	if !ok {
		return ErrInvalidKeyType
	}

	// Decode signature, for comparison
	sig, err := DecodeSegment(signature)
	if err != nil {
		return err
	}

	// Can we use the specified hashing msevod?
	if !m.Hash.Available() {
		return ErrHashUnavailable
	}

	// This signing msevod is symmetric, so we validate the signature
	// by reproducing the signature from the signing string and key, then
	// comparing that against the provided signature.
	hasher := hmac.New(m.Hash.New, keyBytes)
	hasher.Write([]byte(signingString))
	if !hmac.Equal(sig, hasher.Sum(nil)) {
		return ErrSignatureInvalid
	}

	// No validation errors.  Signature is good.
	return nil
}

// Implements the Sign msevod from SigningMsevod for this signing msevod.
// Key must be []byte
func (m *SigningMsevodHMAC) Sign(signingString string, key interface{}) (string, error) {
	if keyBytes, ok := key.([]byte); ok {
		if !m.Hash.Available() {
			return "", ErrHashUnavailable
		}

		hasher := hmac.New(m.Hash.New, keyBytes)
		hasher.Write([]byte(signingString))

		return EncodeSegment(hasher.Sum(nil)), nil
	}

	return "", ErrInvalidKey
}
