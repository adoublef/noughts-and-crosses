package jsonwebkey

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"io"

	"github.com/lestrrat-go/jwx/v2/jwk"
)

func ES256Key(r io.Reader) (jwk.Key, error) {
	if r == nil {
		r = rand.Reader
	}

	raw, err := ecdsa.GenerateKey(elliptic.P256(), r)
	if err != nil {
		return nil, err
	}

	return jwk.FromRaw(raw)
}
