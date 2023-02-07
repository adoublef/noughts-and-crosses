package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	auth "github.com/hyphengolang/noughts-and-crosses/internal/auth/service"
	"github.com/hyphengolang/noughts-and-crosses/internal/conf"
	mail "github.com/hyphengolang/noughts-and-crosses/internal/mailing/service"
	rreg "github.com/hyphengolang/noughts-and-crosses/internal/reg/repository"
	sreg "github.com/hyphengolang/noughts-and-crosses/internal/reg/service"
	"github.com/hyphengolang/noughts-and-crosses/internal/smtp"
	jsonwebtoken "github.com/hyphengolang/noughts-and-crosses/pkg/auth/jwt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/rs/cors"
)

func run() error {
	ctx := context.Background()

	nc, err := nats.Connect(conf.NATSURI, nats.UserJWTAndSeed(conf.NATSToken, conf.NATSSeed))
	if err != nil {
		return err
	}
	defer nc.Close()

	conn, err := pgxpool.New(ctx, conf.DBURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// ping the database
	if err := conn.Ping(ctx); err != nil {
		return err
	}

	mux := chi.NewRouter()
	mux.Use(cors.Default().Handler)

	msv := newMailingService(nc)
	mux.Mount("/mail/v0", msv)

	rsv := newRegService(nc, conn)
	mux.Mount("/registry/v0", rsv)

	asv := newAuthService(nc)
	mux.Mount("/auth/v0", asv)

	log.Println("Listening on port", conf.PORT)
	return http.ListenAndServe(fmt.Sprintf(":%d", conf.PORT), mux)
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

func newRegService(nc *nats.Conn, pg *pgxpool.Pool) *sreg.Service {
	srv := sreg.New(nc, rreg.New(pg))

	return srv
}

func newAuthService(nc *nats.Conn) *auth.Service {
	tk := jsonwebtoken.NewTokenClient()
	srv := auth.New(nc, tk)

	return srv
}
