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

func GenerateToken(ctx context.Context, key jwk.Key, opts ...BuildOption) (Token, error) {
	return Build(key, opts...)
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

// Deprecated
func (t Token) Decode() (jwt.Token, error) {
	return jwt.ParseInsecure(t)
}

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

func ParseToken(token []byte, key jwk.Key) (jwt.Token, error) {
	var sep jwt.SignEncryptParseOption
	switch key := key.(type) {
	case jwk.RSAPublicKey:
		sep = jwt.WithKey(jwa.RS256, key)
	case jwk.ECDSAPublicKey:
		sep = jwt.WithKey(jwa.ES256, key)
	default:
		return nil, errors.New(`unsupported encryption`)
	}

	return jwt.Parse(token, sep)
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

type BuildOption func(*jwt.Builder)

func WithEnd(d time.Duration) BuildOption {
	return func(b *jwt.Builder) {
		b.Expiration(time.Now().Add(d))
	}
}

func WithStart(d time.Duration) BuildOption {
	return func(b *jwt.Builder) {
		b.NotBefore(time.Now().Add(d))
	}
}

func WithID(id string) BuildOption {
	return func(b *jwt.Builder) {
		b.JwtID(id)
	}
}

func WithSubject(sub string) BuildOption {
	return func(b *jwt.Builder) {
		b.Subject(sub)
	}
}

func WithPrivateClaims(claims PrivateClaims) BuildOption {
	return func(b *jwt.Builder) {
		for k, v := range claims {
			b.Claim(k, v)
		}
	}
}

type PrivateClaims map[string]any

// Experimental
func (c PrivateClaims) Append(key string, value any) {
	c[key] = value
}

func Build(key jwk.Key, opts ...BuildOption) ([]byte, error) {
	iss := "https://example.com"
	aud := []string{"https://example.com"}

	b := jwt.NewBuilder().Issuer(iss).Audience(aud).IssuedAt(time.Now())
	for _, opt := range opts {
		opt(b)
	}

	tk, err := b.Build()
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
