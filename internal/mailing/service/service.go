// https://www.convictional.com/blog/go-embed
// https://github.com/convictional/template-embed-example/blob/b3b1e0dfe1e6e38e6ce5e5b6e952f85d881d7311/email/email.go
package service

import (
	"embed"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"

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

func (s *Service) routes() {
	// s.mux.Post("/send", s.handleSend())
}

func (s *Service) listen() {
	s.e.Subscribe(events.EventUserLogin, s.handleSendConfirmation())
}

func (s *Service) handleSendConfirmation() nats.MsgHandler {
	type renderArgs struct {
		Href string
	}

	render, err := smtp.Render(confirmationEmail, "templates/confirmation_email.html")
	if err != nil {
		log.Fatal(err)
	}

	return func(msg *nats.Msg) {
		// EventUserLoginResponse
		var data struct {
			Email string
			Token []byte
		}
		{
			if err := events.Decode(msg.Data, &data); err != nil {
				log.Println(err)
				return
			}
		}

		mail := &smtp.Mail{
			To:   []string{data.Email},
			Subj: "Confirm your email for your account",
		}

		args := &renderArgs{
			Href: fmt.Sprintf("%s/get-started/confirm-email?token=%s", conf.ClientURI, string(data.Token)),
		}

		_ = render(mail, args)

		if err := s.smtp.Send(mail); err != nil {
			log.Printf("sending email error: %v", err)
			return
		}

		log.Printf("email sent to %s", data.Email)
	}
}
