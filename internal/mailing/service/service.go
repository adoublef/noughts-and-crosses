// https://www.convictional.com/blog/go-embed
// https://github.com/convictional/template-embed-example/blob/b3b1e0dfe1e6e38e6ce5e5b6e952f85d881d7311/email/email.go
package service

import (
	"embed"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"

	jwt "github.com/hyphengolang/smtp.google/internal/auth/jwt"
	"github.com/hyphengolang/smtp.google/internal/events"
	"github.com/hyphengolang/smtp.google/internal/smtp"
)

//go:embed templates/confirmation_email.html
var confirmationEmail embed.FS

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
	smtp smtp.Sender
	e    events.Broker
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func New(smtp smtp.Sender, nc *nats.Conn) *Service {
	s := &Service{
		mux:  chi.NewRouter(),
		smtp: smtp,
		e:    events.NewClient(nc),
	}
	go s.listen()
	s.routes()
	return s
}

func (s *Service) listen() {
	s.e.Subscribe(events.EventUserLogin.String(), s.handleSendConfirmation())
}

func (s *Service) handleSendConfirmation() nats.MsgHandler {
	render, err := smtp.Render(confirmationEmail, "templates/confirmation_email.html")
	if err != nil {
		log.Fatal(err)
	}

	newToken := func(msg *nats.Msg) (jwt.Token, error) {
		rawToken, err := s.e.Request(events.EventTokenGenerate.String(), msg.Data, 5*time.Second)
		if err != nil {
			return nil, err
		}

		var token jwt.Token
		return token, s.e.Decode(rawToken, &token)
	}

	type R struct {
		Href string
	}

	return func(msg *nats.Msg) {
		token, err := newToken(msg)
		if err != nil {
			log.Printf("request token error: %v", err)
			return
		}

		// NOTE should handle error here
		html, _ := render("Confirm your email for your account", &R{
			Href: "http://localhost:3000/get-started/confirm-email?token=" + token.String(),
		})

		var m struct {
			Email string
		}
		if err := s.e.Decode(msg.Data, &m); err != nil {
			// return error to producer
			log.Printf("decode error: %v", err)
			return
		}

		if err := s.smtp.Send(html, m.Email); err != nil {
			log.Printf("sending email error: %v", err)
			return
		}

		log.Printf("email sent to %s", m.Email)
	}
}

func (s *Service) routes() {
	// s.mux.Post("/send", s.handleSend())
}
