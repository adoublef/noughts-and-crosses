// https://www.convictional.com/blog/go-embed
// https://github.com/convictional/template-embed-example/blob/b3b1e0dfe1e6e38e6ce5e5b6e952f85d881d7311/email/email.go
package service

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"

	"github.com/hyphengolang/noughts-and-crosses/internal/conf"
	"github.com/hyphengolang/noughts-and-crosses/internal/events"
	"github.com/hyphengolang/noughts-and-crosses/internal/smtp"
)

//go:embed templates/confirmation_login.html
var confirmLogin embed.FS

//go:embed templates/confirmation_signup.html
var confirmSignUp embed.FS

var (
// _, b, _, _ = runtime.Caller(0)
// basepath   = filepath.Dir(b)
)

// func init() {
// 	// _ = tpl

// 	// tpl2 := template.Must(template.ParseFS(confirmationEmail, "templates/confirmation.html"))
// 	// _ = tpl2
// }

var _ http.Handler = (*Service)(nil)

type Service struct {
	mux  chi.Router
	smtp smtp.Mailer
	e    events.Broker
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func New(smtp smtp.Mailer, nc *nats.Conn) *Service {
	s := &Service{
		mux:  chi.NewRouter(),
		smtp: smtp,
		e:    events.NewClient(nc),
	}
	go s.listen()
	s.routes()
	return s
}

func (s *Service) routes() {
	// s.mux.Post("/send", s.handleSend())
}

func (s *Service) listen() {
	s.e.Subscribe(events.EventUserLogin, s.handleLoginConfirmation())
	s.e.Subscribe(events.EventUserSignup, s.handleSignupConfirmation())
}

func (s *Service) handleSignupConfirmation() nats.MsgHandler {
	type renderArgs struct {
		Href string
	}

	render, err := smtp.Render(confirmSignUp, "templates/confirmation_signup.html")
	if err != nil {
		log.Fatal(err)
	}

	newToken := func(msg *nats.Msg) ([]byte, error) {
		raw, err := s.e.Conn().RequestMsg(&nats.Msg{Subject: events.EventTokenGenerateSignUp, Data: msg.Data}, 5*time.Second)
		if err != nil {
			return nil, err
		}

		var tokGen struct {
			Token []byte
		}
		if err := events.Decode(raw.Data, &tokGen); err != nil {
			return nil, err
		}

		return tokGen.Token, nil
	}

	send := func(to string, token []byte) error {
		mail := &smtp.Mail{
			To:   []string{to},
			Subj: "Signup Confirmation",
		}

		args := &renderArgs{
			Href: fmt.Sprintf("%s/signup/confirm-email?token=%s", conf.ClientURI, string(token)),
		}

		if err := render(mail, args); err != nil {
			return err
		}

		return s.smtp.Send(mail)
	}

	return func(msg *nats.Msg) {
		var userIn struct {
			Email    string
			Username string
		}
		if err := events.Decode(msg.Data, &userIn); err != nil {
			log.Println(err)
			return
		}

		tk, err := newToken(msg)
		if err != nil {
			log.Println(err)
			return
		}

		if err := send(userIn.Email, tk); err != nil {
			log.Printf("sending email error: %v", err)
			return
		}

		log.Printf("email sent to %s", userIn.Email)
	}
}

func (s *Service) handleLoginConfirmation() nats.MsgHandler {
	type renderArgs struct {
		Href string
	}

	render, err := smtp.Render(confirmLogin, "templates/confirmation_login.html")
	if err != nil {
		log.Fatal(err)
	}

	send := func(to string, token []byte) error {
		mail := &smtp.Mail{
			To:   []string{to},
			Subj: "Login Confirmation",
		}

		args := &renderArgs{
			Href: fmt.Sprintf("%s/login/confirm-email?token=%s", conf.ClientURI, string(token)),
		}

		if err := render(mail, args); err != nil {
			return err
		}

		return s.smtp.Send(mail)
	}

	return func(msg *nats.Msg) {
		var userIn struct {
			Email string
			Token []byte
		}
		if err := events.Decode(msg.Data, &userIn); err != nil {
			log.Println(err)
			return
		}

		if err := send(userIn.Email, userIn.Token); err != nil {
			log.Printf("sending email error: %v", err)
			return
		}

		log.Printf("email sent to %s", userIn.Email)
	}
}
