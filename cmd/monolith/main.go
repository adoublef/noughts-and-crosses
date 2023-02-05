package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	jsonwebtoken "github.com/hyphengolang/noughts-and-crosses/internal/auth/jwt"
	auth "github.com/hyphengolang/noughts-and-crosses/internal/auth/service"
	"github.com/hyphengolang/noughts-and-crosses/internal/conf"
	mail "github.com/hyphengolang/noughts-and-crosses/internal/mailing/service"
	reg "github.com/hyphengolang/noughts-and-crosses/internal/reg/service"
	"github.com/hyphengolang/noughts-and-crosses/internal/smtp"
	"github.com/nats-io/nats.go"
	"github.com/rs/cors"
)

func run() error {
	fmt.Print(conf.NATSURI)
	nc, err := nats.Connect(nats.DefaultURL, nats.Token(conf.NATSToken))
	if err != nil {
		return err
	}
	defer nc.Close()

	mux := chi.NewRouter()
	mux.Use(cors.Default().Handler)

	msv := newMailingService(nc)
	mux.Mount("/mail/v0", msv)

	rsv := newRegService(nc)
	mux.Mount("/reg/v0", rsv)

	asv := newAuthService(nc)
	mux.Mount("/auth/v0", asv)

	log.Println("Listening on port 8080")
	return http.ListenAndServe("0.0.0.0:8080", mux)
}

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

// example: POST http://localhost:8080/mailing/v0/send
func newMailingService(nc *nats.Conn) *mail.Service {
	// do not hard-code SMTP Port
	e := smtp.NewMailer(conf.SMTPUsername, conf.SMTPPassword, conf.SMTPHost, 587)
	srv := mail.New(e, nc)

	return srv
}

func newRegService(nc *nats.Conn) *reg.Service {
	srv := reg.New(nc)

	return srv
}

func newAuthService(nc *nats.Conn) *auth.Service {
	tk := jsonwebtoken.NewTokenClient()
	srv := auth.New(nc, tk)

	return srv
}
