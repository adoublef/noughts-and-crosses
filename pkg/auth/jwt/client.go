package jsonwebtoken

import (
	"context"
	"io"
	"net/http"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"

	keys "github.com/hyphengolang/noughts-and-crosses/pkg/auth/jwt/jwk"
)

type ClientOption func(*tokenClient)

func WithRandReader(r io.Reader) ClientOption {
	var err error
	return func(c *tokenClient) {
		c.key, err = keys.ES256Key(r)
		if err != nil {
			panic(err)
		}
	}
}

func WithPEM(s string) ClientOption {
	var err error
	return func(c *tokenClient) {
		c.key, err = keys.FromPEM(s)
		if err != nil {
			panic(err)
		}
	}
}

type Client interface {
	ParseToken(token []byte) (jwt.Token, error)
	// ParseRequest parses the request and returns the token
	// if the token is valid, and returns an error otherwise
	ParseRequest(r *http.Request) (jwt.Token, error)
	// ParseCookie parses the cookie and returns the token
	// if the token is valid, and returns an error otherwise
	ParseCookie(r *http.Request, cookieName string) (jwt.Token, error)
	// SignToken generates a token with the given duration
	// and subject. The token is signed with the key provided
	SignToken(ctx context.Context, opts ...BuildOption) ([]byte, error)
}

type tokenClient struct {
	key jwk.Key
}

// ValidateRequest implements TokenClient
func (c *tokenClient) ParseRequest(r *http.Request) (jwt.Token, error) {
	key, err := c.key.PublicKey()
	if err != nil {
		return nil, err
	}

	return ParseRequest(r, key)
}

func (c *tokenClient) ParseToken(token []byte) (jwt.Token, error) {
	key, err := c.key.PublicKey()
	if err != nil {
		return nil, err
	}

	return ParseToken(token, key)
}

func (c *tokenClient) ParseCookie(r *http.Request, cookieName string) (jwt.Token, error) {
	key, err := c.key.PublicKey()
	if err != nil {
		return nil, err
	}

	return ParseCookie(r, key, cookieName)
}

// GenerateToken implements TokenClient
func (c *tokenClient) SignToken(ctx context.Context, opts ...BuildOption) ([]byte, error) {
	return SignToken(ctx, c.key, opts...)
}

func NewTokenClient(opts ...ClientOption) Client {
	c := &tokenClient{}
	for _, opt := range opts {
		opt(c)
	}

	if c.key == nil {
		var err error
		if c.key, err = keys.ES256Key(nil); err != nil {
			panic(err)
		}
	}

	return c
}
