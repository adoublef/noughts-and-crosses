package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hyphengolang/noughts-and-crosses/internal/conf"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/cors"
)

func run() error {
	ctx := context.Background()

	nc, err := newNATService()
	if err != nil {
		return err
	}
	defer nc.Close()

	conn, err := newDBConn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	mux := chi.NewRouter()
	// root
	{
		opt := cors.Options{
			AllowedOrigins:   []string{conf.ClientURI},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		}
		mux.Use(cors.New(opt).Handler)
		mux.Use(middleware.Logger)
		mux.Use(middleware.Recoverer)

		mux.Get("/health", handleHealth(conn))
		mux.Post("/ping", handlePing())
	}

	mux.Mount("/mail", newMailingService(nc))
	mux.Mount("/registry", newRegService(nc, conn))
	mux.Mount("/auth", newAuthService(nc))

	mux.Mount("/games", newGamesService())

	return http.ListenAndServe(fmt.Sprintf(":%d", conf.PORT), mux)
}

func handleHealth(conn *pgxpool.Pool) http.HandlerFunc {
	type response struct {
		Health string `json:"health"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// ping connection
		if err := conn.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"status": "error"}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response{Health: "ok"})
	}
}

func handlePing() http.HandlerFunc {
	// decode the request body into a new `Post` struct
	type request struct {
		Hello string `json:"hello"`
	}

	type response struct {
		Sum float64 `json:"sum"`
		Avg float64 `json:"avg"`
	}
	return func(w http.ResponseWriter, r *http.Request) {

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

		// write the response
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response{Sum: sum, Avg: avg})
	}
}

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}
