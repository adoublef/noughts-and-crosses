package service

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/hyphengolang/noughts-and-crosses/internal/events"
	"github.com/hyphengolang/noughts-and-crosses/internal/service"
	jot "github.com/hyphengolang/noughts-and-crosses/pkg/auth/jwt"
)

var _ http.Handler = (*Service)(nil)

type Service struct {
	m  service.Router
	e  events.Broker
	tk jot.TokenClient
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.m.ServeHTTP(w, r)
}

func New(nc *nats.Conn, tk jot.TokenClient) *Service {
	s := &Service{
		m:  service.NewRouter(),
		e:  events.NewClient(nc),
		tk: tk,
	}
	// go s.listen()
	s.routes()
	return s
}

func (s *Service) routes() {
	s.m.Post("/login", s.handleLogin())
	s.m.Delete("/login", s.handleLogout())
	// should rename to `/login/verify` to
	// avoid confusion or `/token/verify`
	s.m.Get("/login", s.handleVerifyLogin())

	s.m.Get("/token", s.handleRefreshToken())
}

func (s *Service) handleLogin() http.HandlerFunc {
	type loginEmailReq struct {
		Email string `json:"email"`
	}

	type restResp struct {
		Message string `json:"message"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var q loginEmailReq
		if err := s.m.Decode(w, r, &q); err != nil {
			s.m.Respond(w, r, err, http.StatusBadRequest)
			return
		}

		// token creation request
		// could be in a go-routine as its not super necessary to be sequential
		go func() {
			tk, err := s.tk.GenerateToken(context.Background(), jot.WithEnd(5*time.Minute), jot.WithPrivateClaims(jot.PrivateClaims{"email": q.Email}))
			if err != nil {
				// NOTE - send info to logging service
				log.Println(err)
				return
			}

			data := struct {
				Email string
				Token []byte
			}{Email: q.Email, Token: tk}
			p, err := events.Encode(data)
			if err != nil {
				// NOTE - send info to logging service
				log.Println(err)
				return
			}

			if err := s.e.Conn().PublishMsg(&nats.Msg{Subject: events.EventUserLogin, Data: p}); err != nil {
				// NOTE - send info to logging service
				log.Println(err)
				return
			}
		}()

		p := restResp{Message: "check emails for confirmation"}
		s.m.Respond(w, r, p, http.StatusOK)
	}
}

func (s *Service) handleVerifyLogin() http.HandlerFunc {
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
