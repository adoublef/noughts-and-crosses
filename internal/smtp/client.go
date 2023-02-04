package smtp

import (
	"net/smtp"
)

type Sender interface {
	Send(msg []byte, to ...string) error
}

type client struct {
	// SMTP Auth
	a smtp.Auth
	// Username
	u string
	// Host
	h string
}

func (e *client) Send(msg []byte, to ...string) error {
	return smtp.SendMail(e.h+":587", e.a, e.u, to, msg)
}

func (e *client) SendTLS(msg []byte, to ...string) error {
	return smtp.SendMail(e.h+":465", e.a, e.u, to, msg)
}

func NewClient(host, username, password string) Sender {
	e := &client{
		a: smtp.PlainAuth("", username, password, host),
		u: username,
		h: host,
	}

	return e
}
