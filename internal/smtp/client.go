// https://golang.hotexamples.com/examples/net.smtp/-/NewClient/golang-newclient-function-examples.html
// https://gist.github.com/jpillora/cb46d183eca0710d909a
// https://gist.github.com/andelf/5118732
package smtp

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
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

// func (e *client) SendTLS(msg []byte, to ...string) error {
// 	return smtp.SendMail(e.h+":465", e.a, e.u, to, msg)
// }

func NewSender(host, username, password string) Sender {
	e := &client{
		a: smtp.PlainAuth("", username, password, host),
		u: username,
		h: host,
	}

	return e
}

type Mailer interface {
	Send(m *Mail) error
}

type mailClient struct {
	usern, passw, host string
	port               int
}

func NewMailer(usern, passw, host string, port int) Mailer {
	m := &mailClient{usern: usern, passw: passw, host: host, port: port}
	return m
}

func (mc *mailClient) Send(m *Mail) error {
	addr := fmt.Sprintf("%s:%d", mc.host, mc.port)
	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer c.Close()

	if err = c.StartTLS(&tls.Config{InsecureSkipVerify: true, ServerName: mc.host}); err != nil {
		return err
	}

	a := smtp.PlainAuth("", mc.usern, mc.passw, mc.host)
	if err := c.Auth(a); err != nil {
		return err
	}

	if err := c.Mail(mc.usern); err != nil {
		return err
	}

	// does not currently support bcc
	for _, r := range append(m.To, m.CC...) {
		if err := c.Rcpt(r); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	var sb bytes.Buffer
	{
		sb.WriteString(fmt.Sprintf("From: %s", mc.usern))
		sb.WriteString("\r\n")
		if len(m.To) > 0 {
			sb.WriteString(fmt.Sprintf("To: %s", strings.Join(m.To, ",")))
			sb.WriteString("\r\n")
		}
		if len(m.CC) > 0 {
			sb.WriteString(fmt.Sprintf("Cc: %s", strings.Join(m.CC, ",")))
			sb.WriteString("\r\n")
		}
		if len(m.Subj) > 0 {
			sb.WriteString(fmt.Sprintf("Subject: %s", m.Subj))
			sb.WriteString("\r\n")
		}
		hdr := []string{"MIME-version: 1.0;", "Content-Type: text/html; charset=\"UTF-8\";"}
		sb.WriteString(strings.Join(hdr, "\r\n"))
		sb.WriteString("\r\n")
		sb.WriteString("\r\n")
		sb.Write(m.Body)
	}
	msg := sb.Bytes()
	if _, err = w.Write(msg); err != nil {
		return err
	}

	if err = w.Close(); err != nil {
		return err
	}

	return c.Quit()
}

type Mail struct {
	From string
	To   []string
	CC   []string
	Subj string
	Body []byte
}
