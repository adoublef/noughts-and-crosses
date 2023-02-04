package jsonwebtoken_test

import (
	"context"
	"testing"
	"time"

	"github.com/hyphengolang/prelude/testing/is"
	jot "github.com/hyphengolang/smtp.google/internal/auth/jwt"
	jok "github.com/hyphengolang/smtp.google/internal/auth/jwt/jwk"
)

func TestGenerateToken(t *testing.T) {
	is := is.New(t)

	key, err := jok.ES256Key(nil)
	is.NoErr(err) // unable to generate jwt key from raw

	_, err = jot.GenerateToken(context.Background(), key, 0, time.Minute, "foo@mail.com")
	is.NoErr(err) // unable to generate token
}
