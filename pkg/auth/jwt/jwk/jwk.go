package jsonwebkey

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"io"

	"github.com/lestrrat-go/jwx/v2/jwk"
)

// https://pawamoy.github.io/posts/pass-makefile-args-as-typed-in-command-line/
// https://www.cyberciti.biz/faq/linux-unix-formatting-dates-for-display/

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

func FromPEM(s string) (jwk.Key, error) {
	return jwk.ParseKey([]byte(s), jwk.WithPEM(true))
}
