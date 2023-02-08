package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	auth "github.com/hyphengolang/noughts-and-crosses/internal/auth/service"
	"github.com/hyphengolang/noughts-and-crosses/internal/conf"
	"github.com/hyphengolang/noughts-and-crosses/internal/events"
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

	mux.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{conf.ClientURI},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}).Handler)
	{
		mux.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"status": "get"}`))
		})

		mux.Post("/health", func(w http.ResponseWriter, r *http.Request) {
			// decode the request body into a new `Post` struct
			type request struct {
				Hello string `json:"hello"`
			}

			var body request
			err := json.NewDecoder(r.Body).Decode(&body)

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"status": "error"}`))
				return
			}

			// find the sum of the all letters in Hello
			sum := 0.0
			for _, c := range body.Hello {
				sum += float64(int(c) - 32)
			}

			avg := sum / float64(len(body.Hello))

			type response struct {
				Sum float64 `json:"sum"`
				Avg float64 `json:"avg"`
			}

			// write the response
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response{Sum: sum, Avg: avg})
		})
	}
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

func newMailingService(nc *nats.Conn) *mail.Service {
	em := smtp.NewMailer(conf.SMTPUsername, conf.SMTPPassword, conf.SMTPHost, 587)
	ec := events.NewClient(nc)
	return mail.New(em, ec)
}

func newRegService(nc *nats.Conn, pg *pgxpool.Pool) *sreg.Service {
	ec := events.NewClient(nc)
	return sreg.New(ec, rreg.New(pg))
}

func newAuthService(nc *nats.Conn) *auth.Service {
	tk := jsonwebtoken.NewTokenClient()
	ec := events.NewClient(nc)
	return auth.New(ec, tk)
}
