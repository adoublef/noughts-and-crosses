package events

import "github.com/nats-io/nats.go"

type Event string

const (
	EventLoginConfirmation   = "user.email.login"
	EventSignupConfirmation  = "user.email.signup"
	EventGenerateLoginToken  = "token.generate.login" // NOTE not used
	EventGenerateSignUpToken = "token.generate.signup"
	// EventTokenVerify         = "token.verify"
)

// TODO implement Unmarshaler and Marshaler interfaces for binary encoding
type DataSignUpConfirm struct {
	Email    string
	Username string
}

// TODO implement Unmarshaler and Marshaler interfaces for binary encoding
type DataLoginConfirm struct {
	Email string
	Token []byte
}

// TODO implement Unmarshaler and Marshaler interfaces for binary encoding
type DataTokenGen struct {
	Token []byte
}

func NewSignupTokenMsg(token []byte) (*nats.Msg, error) {
	v := DataTokenGen{Token: token}
	p, err := Encode(v)
	if err != nil {
		return nil, err
	}
	// Request from Auth service to get token from header.
	msg := nats.Msg{
		Subject: EventSignupConfirmation,
		// token from header
		Data: p,
	}
	return &msg, nil
}

func NewSignupConfirmationMsg(email, username string) (*nats.Msg, error) {
	// send email to complete sign-up process
	// automatically check which email provider so
	// can send a link to the correct email provider
	// https://www.freecodecamp.org/news/the-best-free-email-providers-2021-guide-to-online-email-account-services/
	data := DataSignUpConfirm{Email: email, Username: username}
	p, err := Encode(data)
	if err != nil {
		return nil, err
	}

	return &nats.Msg{Subject: EventSignupConfirmation, Data: p}, nil
}

func NewLoginConfirmationMsg(email string, token []byte) (*nats.Msg, error) {
	data := DataLoginConfirm{Email: email, Token: token}
	p, err := Encode(data)
	if err != nil {
		return nil, err
	}

	return &nats.Msg{Subject: EventLoginConfirmation, Data: p}, nil
}
