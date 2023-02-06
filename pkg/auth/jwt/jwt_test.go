package jsonwebtoken_test

import (
	"context"
	"testing"
	"time"

	jot "github.com/hyphengolang/noughts-and-crosses/pkg/auth/jwt"
	jok "github.com/hyphengolang/noughts-and-crosses/pkg/auth/jwt/jwk"
	"github.com/hyphengolang/prelude/testing/is"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

func TestGenerateToken(t *testing.T) {
	is := is.New(t)

	key, err := jok.ES256Key(nil)
	is.NoErr(err) // unable to generate jwt key from raw

	_, err = jot.GenerateToken(context.Background(), key, jot.WithEnd(time.Minute), jot.WithSubject("foo@mail.com"))
	is.NoErr(err) // unable to generate token
}

func TestToken(t *testing.T) {
	is := is.New(t)

	// token, err := Build(jwk.ES256(), WithStart(time.Minute), WithEnd(time.Minute), WithID("foo"))

	var key jwk.Key
	token, err := jot.Build(key,
		jot.WithEnd(time.Minute),
		jot.WithPrivateClaims(jot.PrivateClaims{
			"email": "test@mail.com",
		}),
	)

	is.NoErr(err)         // generate token
	is.True(token != nil) // token is nil
}
