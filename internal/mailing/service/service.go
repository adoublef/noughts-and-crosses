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
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"

	jwt "github.com/hyphengolang/noughts-and-crosses/internal/auth/jwt"
	"github.com/hyphengolang/noughts-and-crosses/internal/conf"
	"github.com/hyphengolang/noughts-and-crosses/internal/events"
	"github.com/hyphengolang/noughts-and-crosses/internal/smtp"
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

func (s *Service) listen() {
	s.e.Subscribe(events.EventUserLogin, s.handleSendConfirmation())
}

func (s *Service) handleSendConfirmation() nats.MsgHandler {
	render, err := smtp.Render(confirmationEmail, "templates/confirmation_email.html")
	if err != nil {
		log.Fatal(err)
	}

	type Q struct {
		Email string
	}

	type email struct {
		Href string
	}

	newToken := func(q Q) (jwt.Token, error) {
		rawToken, err := s.e.Request(events.EventTokenGenerate, q, 5*time.Second)
		if err != nil {
			return nil, err
		}

		var token jwt.Token
		return token, s.e.Decode(rawToken, &token)
	}

	return func(msg *nats.Msg) {
		var q Q
		if err := s.e.Decode(msg.Data, &q); err != nil {
			// return error to producer
			log.Printf("mailing.service: decode error: %v", err)
			return
		}

		// generate token
		token, err := newToken(q)
		if err != nil {
			log.Printf("request token error: %v", err)
			return
		}

		mail := &smtp.Mail{
			To:   []string{q.Email},
			Subj: "Confirm your email for your account: " + uuid.New().String(),
			// Body: html,
		}

		href := fmt.Sprintf("%s/get-started/confirm-email?token=%s", conf.ClientURI, token.String())
		_ = render(mail, &email{Href: href})

		if err := s.smtp.Send(mail); err != nil {
			log.Printf("sending email error: %v", err)
			return
		}

		log.Printf("email sent to %s", q.Email)
	}
}

func (s *Service) routes() {
	// s.mux.Post("/send", s.handleSend())
}
