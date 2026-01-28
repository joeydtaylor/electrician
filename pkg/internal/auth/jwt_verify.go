package auth

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
)

func verifySignature(alg, signingInput string, sig []byte, key interface{}) error {
	switch alg {
	case "RS256":
		return verifyRSAPKCS1v15(crypto.SHA256, signingInput, sig, key)
	case "RS384":
		return verifyRSAPKCS1v15(crypto.SHA384, signingInput, sig, key)
	case "RS512":
		return verifyRSAPKCS1v15(crypto.SHA512, signingInput, sig, key)
	case "PS256":
		return verifyRSAPSS(crypto.SHA256, signingInput, sig, key)
	case "PS384":
		return verifyRSAPSS(crypto.SHA384, signingInput, sig, key)
	case "PS512":
		return verifyRSAPSS(crypto.SHA512, signingInput, sig, key)
	case "ES256":
		return verifyECDSA(crypto.SHA256, signingInput, sig, key)
	case "ES384":
		return verifyECDSA(crypto.SHA384, signingInput, sig, key)
	case "ES512":
		return verifyECDSA(crypto.SHA512, signingInput, sig, key)
	default:
		return fmt.Errorf("unsupported jwt alg %q", alg)
	}
}

func verifyRSAPKCS1v15(hash crypto.Hash, signingInput string, sig []byte, key interface{}) error {
	pub, ok := rsaPublicKey(key)
	if !ok {
		return errors.New("rsa public key required")
	}
	hashed, err := hashBytes(hash, signingInput)
	if err != nil {
		return err
	}
	return rsa.VerifyPKCS1v15(pub, hash, hashed, sig)
}

func verifyRSAPSS(hash crypto.Hash, signingInput string, sig []byte, key interface{}) error {
	pub, ok := rsaPublicKey(key)
	if !ok {
		return errors.New("rsa public key required")
	}
	hashed, err := hashBytes(hash, signingInput)
	if err != nil {
		return err
	}
	return rsa.VerifyPSS(pub, hash, hashed, sig, &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash, Hash: hash})
}

func verifyECDSA(hash crypto.Hash, signingInput string, sig []byte, key interface{}) error {
	pub, ok := key.(*ecdsa.PublicKey)
	if !ok {
		return errors.New("ecdsa public key required")
	}
	hashed, err := hashBytes(hash, signingInput)
	if err != nil {
		return err
	}
	keyBytes := (pub.Curve.Params().BitSize + 7) / 8
	r, s, err := parseECDSASignature(sig, keyBytes)
	if err != nil {
		return err
	}
	if !ecdsa.Verify(pub, hashed, r, s) {
		return errors.New("ecdsa signature invalid")
	}
	return nil
}

func parseECDSASignature(sig []byte, keyBytes int) (*big.Int, *big.Int, error) {
	if len(sig) != 2*keyBytes {
		return nil, nil, fmt.Errorf("invalid ecdsa signature length %d", len(sig))
	}
	r := new(big.Int).SetBytes(sig[:keyBytes])
	s := new(big.Int).SetBytes(sig[keyBytes:])
	return r, s, nil
}

func hashBytes(hash crypto.Hash, input string) ([]byte, error) {
	switch hash {
	case crypto.SHA256:
		sum := sha256.Sum256([]byte(input))
		return sum[:], nil
	case crypto.SHA384:
		sum := sha512.Sum384([]byte(input))
		return sum[:], nil
	case crypto.SHA512:
		sum := sha512.Sum512([]byte(input))
		return sum[:], nil
	default:
		if !hash.Available() {
			return nil, errors.New("hash unavailable")
		}
		h := hash.New()
		_, _ = h.Write([]byte(input))
		return h.Sum(nil), nil
	}
}

func decodeJWTPart(part string) ([]byte, error) {
	if part == "" {
		return nil, errors.New("empty jwt part")
	}
	return base64.RawURLEncoding.DecodeString(part)
}
