package service

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"

	jwt "github.com/hyphengolang/smtp.google/internal/auth/jwt"
	"github.com/hyphengolang/smtp.google/internal/events"
	"github.com/hyphengolang/smtp.google/internal/service"
)

var _ http.Handler = (*Service)(nil)

type Service struct {
	m  service.Router
	e  events.Broker
	tk jwt.TokenClient
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.m.ServeHTTP(w, r)
}

func New(nc *nats.Conn, tk jwt.TokenClient) *Service {
	s := &Service{
		m:  service.NewRouter(),
		e:  events.NewClient(nc),
		tk: tk,
	}
	go s.listen()
	s.routes()
	return s
}

func (s *Service) routes() {
	s.m.Post("/login", s.handleLogin())
	s.m.Get("/login", s.handleVerify())
	s.m.Delete("/login", s.handleLogout())

	s.m.Get("/token", s.handleRefreshToken())
}

func (s *Service) handleLogin() http.HandlerFunc {
	type Q struct {
		Email string `json:"email"`
	}

	type P struct {
		Message string `json:"message"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var q Q
		if err := s.m.Decode(w, r, &q); err != nil {
			s.m.Respond(w, r, err, http.StatusBadRequest)
			return
		}

		if err := s.e.Publish(events.EventUserLogin.String(), q); err != nil {
			s.m.Respond(w, r, err, http.StatusInternalServerError)
			return
		}

		p := P{Message: "check emails for confirmation"}
		s.m.Respond(w, r, p, http.StatusOK)
	}
}

func (s *Service) handleVerify() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := s.tk.ParseRequest(r)
		if err != nil {
			s.m.Respond(w, r, err, http.StatusUnauthorized)
			return
		}

		s.m.Respond(w, r, nil, http.StatusOK)
	}
}

func (s *Service) handleLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.m.Respond(w, r, nil, http.StatusNotImplemented)
	}
}

func (s *Service) handleRefreshToken() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.m.Respond(w, r, nil, http.StatusNotImplemented)
	}
}

func (s *Service) listen() {
	s.e.Subscribe(events.EventTokenGenerate.String(), s.handleTokenGenerate())
}

func (s *Service) handleTokenGenerate() nats.MsgHandler {
	return func(msg *nats.Msg) {
		tk, err := s.tk.GenerateToken(context.Background(), 30*time.Second, uuid.New().String())
		if err != nil {
			log.Println(err)
			return
		}

		if err := s.e.Publish(msg.Reply, tk); err != nil {
			log.Println(err)
			return
		}

		log.Println("token generated")
	}
}
