package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

func parseJWK(k jwkKey) (interface{}, error) {
	switch strings.ToUpper(k.Kty) {
	case "RSA":
		return parseRSAPublicKey(k.N, k.E)
	case "EC":
		return parseECPublicKey(k.Crv, k.X, k.Y)
	default:
		return nil, fmt.Errorf("unsupported jwk kty %q", k.Kty)
	}
}

func parseRSAPublicKey(nStr, eStr string) (*rsa.PublicKey, error) {
	if nStr == "" || eStr == "" {
		return nil, errors.New("rsa jwk missing n/e")
	}
	nb, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, err
	}
	eb, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, err
	}
	e := 0
	for _, b := range eb {
		e = (e << 8) + int(b)
	}
	if e == 0 {
		return nil, errors.New("rsa jwk invalid exponent")
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(nb), E: e}, nil
}

func parseECPublicKey(crv, xStr, yStr string) (*ecdsa.PublicKey, error) {
	if crv == "" || xStr == "" || yStr == "" {
		return nil, errors.New("ec jwk missing crv/x/y")
	}
	var curve elliptic.Curve
	switch crv {
	case "P-256":
		curve = elliptic.P256()
	case "P-384":
		curve = elliptic.P384()
	case "P-521":
		curve = elliptic.P521()
	default:
		return nil, fmt.Errorf("unsupported ec curve %q", crv)
	}
	xb, err := base64.RawURLEncoding.DecodeString(xStr)
	if err != nil {
		return nil, err
	}
	yb, err := base64.RawURLEncoding.DecodeString(yStr)
	if err != nil {
		return nil, err
	}
	x := new(big.Int).SetBytes(xb)
	y := new(big.Int).SetBytes(yb)
	if !curve.IsOnCurve(x, y) {
		return nil, errors.New("ec jwk point not on curve")
	}
	return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}, nil
}
