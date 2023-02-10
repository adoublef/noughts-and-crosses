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

var (
	//go:embed templates/confirmation_login.html
	confirmLogin embed.FS

	//go:embed templates/confirmation_signup.html
	confirmSignUp embed.FS
)

type Service struct {
	mux  chi.Router
	smtp smtp.Mailer
	e    events.Broker
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func New(smtp smtp.Mailer, e events.Broker) *Service {
	s := &Service{
		mux:  chi.NewRouter(),
		smtp: smtp,
		e:    e,
	}
	go s.listen()
	s.routes()
	return s
}

func (s *Service) routes() {
	// s.mux.Post("/send", s.handleSend())
}

func (s *Service) listen() {
	// response needed from `auth`
	s.e.Subscribe(events.EventSendLoginConfirm, s.handleLoginConfirm())
	s.e.Subscribe(events.EventSendSignupConfirm, s.handleSignupConfirm())
}

func (s *Service) handleSignupConfirm() nats.MsgHandler {
	type Args struct {
		Href string
	}

	render, err := smtp.Render(confirmSignUp, "templates/confirmation_signup.html")
	if err != nil {
		log.Fatalf("render confirmation signup: %v", err)
	}

	send := func(to string, token []byte) error {
		args := &Args{
			Href: fmt.Sprintf("%s/signup/confirm-email?token=%s", conf.ClientURI, string(token)),
		}

		mail, err := render(args, "Signup Confirmation", to)
		if err != nil {
			return err
		}

		return s.smtp.Send(mail)
	}

	request := func(msg *nats.Msg) (token []byte, err error) {
		type Data struct{ events.Data[[]byte] }

		msg = events.Redirect(events.EventGenerateSignupToken, msg)
		raw, err := s.e.Request(msg, 5*time.Second)
		if err != nil {
			return
		}

		var reply Data
		if err := events.Unmarshal(raw, &reply); err != nil {
			return nil, err
		}

		return reply.Value, reply.Err
	}

	return func(msg *nats.Msg) {
		var in events.DataSignUpConfirm
		if err := events.Unmarshal(msg.Data, &in); err != nil {
			log.Printf("gob decoding: %v", err)
			return
		}

		token, err := request(msg)
		if err != nil {
			log.Printf("parsing token: %v", err)
			return
		}

		if err := send(in.Email, token); err != nil {
			log.Printf("sending email: %v", err)
			return
		}
	}
}

func (s *Service) handleLoginConfirm() nats.MsgHandler {
	type Args struct {
		Href string
	}

	render, err := smtp.Render(confirmLogin, "templates/confirmation_login.html")
	if err != nil {
		log.Fatalf("render confirmation login: %v", err)
	}

	send := func(to string, token []byte) error {
		args := &Args{
			Href: fmt.Sprintf("%s/login/confirm-email?token=%s", conf.ClientURI, string(token)),
		}

		mail, err := render(args, "Login Confirmation", to)
		if err != nil {
			return err
		}

		return s.smtp.Send(mail)
	}

	return func(msg *nats.Msg) {
		var userIn events.DataLoginConfirm
		if err := events.Unmarshal(msg.Data, &userIn); err != nil {
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
