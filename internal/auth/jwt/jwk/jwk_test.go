package jsonwebkey_test

import (
	"testing"

	"github.com/hyphengolang/prelude/testing/is"
	jok "github.com/hyphengolang/smtp.google/internal/auth/jwt/jwk"
)

func TestGeneratePair(t *testing.T) {
	is := is.New(t)

	_, err := jok.ES256Key(nil)
	is.NoErr(err) // unable to generate jwt key from raw
}
