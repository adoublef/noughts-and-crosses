package events

import (
	"bytes"
	"encoding/gob"

	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/nats-io/nats.go"
)

func Decode(p []byte, v any) error {
	return gob.NewDecoder(bytes.NewReader(p)).Decode(v)
}

func Encode(v any) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(v); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type Event string

const (
	EventSendLoginConfirm        = "user.email.login"
	EventSendSignupConfirm       = "user.email.signup"
	EventGenerateLoginToken      = "token.generate.login" // NOTE not used
	EventGenerateSignupToken     = "token.generate.signup"
	EventVerifySignupToken       = "token.verify.signup"
	EventCreateProfileValidation = "token.decode"
)

type DataJWTToken struct {
	Token jwt.Token
}

// TODO implement Unmarshaler and Marshaler interfaces for binary encoding
type DataSignUpConfirm struct {
	Email string
}

// TODO implement Unmarshaler and Marshaler interfaces for binary encoding
type DataLoginConfirm struct {
	Email string
	Token []byte
}

// TODO implement Unmarshaler and Marshaler interfaces for binary encoding
type DataEmailToken struct {
	Token []byte
}

type DataAuthToken struct {
	Token []byte
	Email string
}

// TODO implement Error interface

func NewCreateProfileValidationMsg(email string, token []byte) (*nats.Msg, error) {
	v := DataAuthToken{Token: token, Email: email}
	p, err := Encode(v)
	if err != nil {
		return nil, err
	}

	return &nats.Msg{Subject: EventCreateProfileValidation, Data: p}, nil
}

func NewSignupVerifyMsg(token []byte) (*nats.Msg, error) {
	v := DataEmailToken{Token: token}
	p, err := Encode(v)
	if err != nil {
		return nil, err
	}
	// Request from Auth service to get token from header.
	return &nats.Msg{Subject: EventVerifySignupToken, Data: p}, nil
}

func NewSendSignupConfirmMsg(email string) (*nats.Msg, error) {
	// send email to complete sign-up process
	// automatically check which email provider so
	// can send a link to the correct email provider
	// https://www.freecodecamp.org/news/the-best-free-email-providers-2021-guide-to-online-email-account-services/
	data := DataSignUpConfirm{Email: email}
	p, err := Encode(data)
	if err != nil {
		return nil, err
	}

	return &nats.Msg{Subject: EventSendSignupConfirm, Data: p}, nil
}

func NewLoginConfirmMsg(email string, token []byte) (*nats.Msg, error) {
	data := DataLoginConfirm{Email: email, Token: token}
	p, err := Encode(data)
	if err != nil {
		return nil, err
	}

	return &nats.Msg{Subject: EventSendLoginConfirm, Data: p}, nil
}
