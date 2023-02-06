package jsonwebkey_test

import (
	"testing"

	jok "github.com/hyphengolang/noughts-and-crosses/internal/auth/jwt/jwk"
	"github.com/hyphengolang/prelude/testing/is"
)

func TestGeneratePair(t *testing.T) {
	is := is.New(t)

	_, err := jok.ES256Key(nil)
	is.NoErr(err) // unable to generate jwt key from raw
}
