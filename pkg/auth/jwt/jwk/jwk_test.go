package jsonwebkey_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	keys "github.com/hyphengolang/noughts-and-crosses/pkg/auth/jwt/jwk"
	"github.com/hyphengolang/prelude/testing/is"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

func TestGeneratePair(t *testing.T) {
	is := is.New(t)

	_, err := keys.ES256Key(nil)
	is.NoErr(err) // unable to generate jwt key from raw
}

// NOTE FOR EXAMPLE ONLY, THIS KEY IS NOT TO BE USED IN PRODUCTION

func TestGenerateFromSecret(t *testing.T) {
	is := is.New(t)

	t.Run("using raw", func(t *testing.T) {
		raw, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		is.NoErr(err) // unable to generate jwt key from raw

		key, err := jwk.FromRaw(raw)
		is.NoErr(err) // unable to generate jwt key from raw

		is.Equal(key.KeyType().String(), "EC") // key type is not EC
	})

	t.Run("using env var", func(t *testing.T) {
		private := "-----BEGIN EC PRIVATE KEY-----\nMHcCAQEEIJTOquEY44KTYC5wtBmj+/2XDLT5Q5/b9QxWJ1MN5K+OoAoGCCqGSM49\nAwEHoUQDQgAEyVsiu8IG5oQiNwYX2F7Wh6XD2dMKOTerOmo1YL08O8mMIGyw9qQo\naauid5eOBuv7CtF3bQ6QvsEf6TuiPZwIqQ==\n-----END EC PRIVATE KEY-----"

		k1, err := jwk.ParseKey([]byte(private), jwk.WithPEM(true))
		is.NoErr(err) // unable to generate jwt key from raw

		is.Equal(k1.KeyType().String(), "EC") // key type is not EC

		k2, err := keys.FromPEM(private)
		is.NoErr(err) // unable to generate jwt key from raw

		is.Equal(k1, k2) // keys are not equal
	})
}
