// https://www.convictional.com/blog/go-embed
// https://github.com/convictional/template-embed-example/blob/b3b1e0dfe1e6e38e6ce5e5b6e952f85d881d7311/email/email.go
package service

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/hyphengolang/noughts-and-crosses/internal/conf"
	"github.com/hyphengolang/noughts-and-crosses/internal/events"
	"github.com/hyphengolang/noughts-and-crosses/internal/service"
	"github.com/hyphengolang/noughts-and-crosses/internal/smtp"
)

var (
	//go:embed templates/confirmation_login.html
	confirmLogin embed.FS

	//go:embed templates/confirmation_signup.html
	confirmSignUp embed.FS
)

type Service struct {
	m    service.Router
	smtp smtp.Mailer
	e    events.Broker
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.m.ServeHTTP(w, r)
}

func New(smtp smtp.Mailer, e events.Broker) *Service {
	s := &Service{
		m:    service.NewRouter(),
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
	s.e.Conn().Subscribe(events.EventSendLoginConfirm, s.handleLoginConfirm())
	s.e.Conn().Subscribe(events.EventSendSignupConfirm, s.handleSignupConfirm())
}

func (s *Service) handleSignupConfirm() nats.Handler {
	type Args struct {
		Href string
	}

	render, err := smtp.Render(confirmSignUp, "templates/confirmation_signup.html")
	if err != nil {
		log.Fatalf("render confirmation signup: %v", err)
	}

	send := func(to string, token []byte) error {
		args := &Args{
			Href: fmt.Sprintf("%s/signup/confirm-email?token=%s", s.m.ClientURI(), string(token)),
		}

		mail, err := render(args, "Signup Confirmation", to)
		if err != nil {
			return err
		}

		return s.smtp.Send(mail)
	}

	parseToken := func(msg *events.DataEmail) (email string, token []byte, err error) {
		type Data struct{ events.Data[[]byte] }
		var response Data

		err = s.e.Conn().Request(events.EventGenerateSignupToken, msg, &response, 5*time.Second)
		if err != nil {
			return
		}

		if err = response.Err; err != nil {
			return
		}

		return msg.Email, response.Value, response.Err
	}

	return func(msg *events.DataEmail) {
		_, token, err := parseToken(msg)
		if err != nil {
			log.Printf("request result: %v", err)
			return
		}

		if err := send(msg.Email, token); err != nil {
			log.Printf("sending email: %v", err)
			return
		}
	}
}

func (s *Service) handleLoginConfirm() nats.Handler {
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

	return func(msg *events.DataLoginConfirm) {
		if err := send(msg.Email, msg.Token); err != nil {
			log.Printf("sending login email: %v", err)
			return
		}
	}
}
