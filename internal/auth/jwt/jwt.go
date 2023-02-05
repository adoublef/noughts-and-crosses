package jsonwebtoken

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

var (
	iss = "http://server.example.com"
	aud = []string{"client_a"}
)

func GenerateToken(ctx context.Context, key jwk.Key, start, end time.Duration, subject string) (Token, error) {
	tk, err := jwt.NewBuilder().
		Issuer(iss).
		IssuedAt(time.Now()).
		NotBefore(time.Now().Add(start)).
		Expiration(time.Now().Add(end)).
		Subject(subject).
		Audience(aud).
		// NOTE - per device authentication
		// JwtID(uuid.New().String()).
		Build()

	if err != nil {
		return nil, err
	}

	var sep jwt.SignEncryptParseOption
	switch key := key.(type) {
	case jwk.RSAPrivateKey:
		sep = jwt.WithKey(jwa.RS256, key)
	case jwk.ECDSAPrivateKey:
		sep = jwt.WithKey(jwa.ES256, key)
	default:
		return nil, errors.New(`unsupported encryption`)
	}

	return jwt.Sign(tk, sep)
}

// ParseRequest parses the request and returns the token
// key must be a private key
func Sign(tk jwt.Token, key jwk.Key) (Token, error) {
	var sep jwt.SignEncryptParseOption
	switch key := key.(type) {
	case jwk.RSAPrivateKey:
		sep = jwt.WithKey(jwa.RS256, key)
	case jwk.ECDSAPrivateKey:
		sep = jwt.WithKey(jwa.ES256, key)
	default:
		return nil, errors.New(`unsupported encryption`)
	}

	return jwt.Sign(tk, sep)
}

type Token []byte

func (t Token) String() string { return string(t) }

// ParseRequest parses the request and returns the token
// key must be a public key
func ParseRequest(r *http.Request, key jwk.Key) (jwt.Token, error) {
	var sep jwt.SignEncryptParseOption
	switch key := key.(type) {
	case jwk.RSAPublicKey:
		sep = jwt.WithKey(jwa.RS256, key)
	case jwk.ECDSAPublicKey:
		sep = jwt.WithKey(jwa.ES256, key)
	default:
		return nil, errors.New(`unsupported encryption`)
	}

	return jwt.ParseRequest(r, sep)
}

func ParseCookie(r *http.Request, key jwk.Key, cookieName string) (jwt.Token, error) {
	c, err := r.Cookie(cookieName)
	if err != nil {
		return nil, err
	}

	var sep jwt.SignEncryptParseOption
	switch key := key.(type) {
	case jwk.RSAPublicKey:
		sep = jwt.WithKey(jwa.RS256, key)
	case jwk.ECDSAPublicKey:
		sep = jwt.WithKey(jwa.ES256, key)
	default:
		return nil, errors.New(`unsupported encryption`)
	}

	return jwt.Parse([]byte(c.Value), sep)
}
