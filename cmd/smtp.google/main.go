package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	jsonwebtoken "github.com/hyphengolang/smtp.google/internal/auth/jwt"
	auth "github.com/hyphengolang/smtp.google/internal/auth/service"
	"github.com/hyphengolang/smtp.google/internal/conf"
	mail "github.com/hyphengolang/smtp.google/internal/mailing/service"
	reg "github.com/hyphengolang/smtp.google/internal/reg/service"
	"github.com/hyphengolang/smtp.google/internal/smtp"
	"github.com/nats-io/nats.go"
	"github.com/rs/cors"
)

func run() error {
	nc, err := nats.Connect(conf.NATSURI)
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
	return http.ListenAndServe(":8080", mux)
}

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

// example: POST http://localhost:8080/mailing/v0/send
func newMailingService(nc *nats.Conn) *mail.Service {
	e := smtp.NewClient(conf.SMTPHost, conf.SMTPUsername, conf.SMTPPassword)
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
