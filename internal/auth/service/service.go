package service

import (
	"context"
	"fmt"
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
	go s.listen()
	s.routes()
	return s
}

func (s *Service) routes() {
	s.m.Post("/login", s.handleLogin())
	// s.m.Delete("/login", s.handleLogout())
	// should rename to `/login/verify` to
	// avoid confusion or `/token/verify`
	s.m.Get("/login", s.handleConfirmLogin())

	// s.m.Get("/token", s.handleRefreshToken())
}

func (s *Service) handleLogin() http.HandlerFunc {
	type request struct {
		Email string `json:"email"`
	}

	type response struct {
		Message string `json:"message"`
	}

	newLoginTokenMsg := func(email string) (*nats.Msg, error) {
		tk, err := s.tk.GenerateToken(context.Background(), jot.WithEnd(5*time.Minute), jot.WithPrivateClaims(jot.PrivateClaims{"email": email}))
		if err != nil {
			return nil, err
		}

		data := struct {
			Email string
			Token []byte
		}{Email: email, Token: tk}
		p, err := events.Encode(data)
		if err != nil {
			return nil, err
		}

		return &nats.Msg{Subject: events.EventUserLogin, Data: p}, nil
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var q request
		if err := s.m.Decode(w, r, &q); err != nil {
			s.m.Respond(w, r, err, http.StatusBadRequest)
			return
		}

		msg, err := newLoginTokenMsg(q.Email)
		if err != nil {
			s.m.Respond(w, r, err, http.StatusInternalServerError)
			return
		}

		if err := s.e.Conn().PublishMsg(msg); err != nil {
			s.m.Respond(w, r, err, http.StatusInternalServerError)
			return
		}

		s.m.Respond(w, r, response{
			Message: "check emails for confirmation",
		}, http.StatusOK)
	}
}

func (s *Service) handleConfirmLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tk, err := s.tk.ParseRequest(r)
		if err != nil {
			s.m.Respond(w, r, err, http.StatusUnauthorized)
			return
		}

		// ask registry service if user exists
		{
			fmt.Println(tk.PrivateClaims()["email"].(string))

			// email, _ := tk.PrivateClaims()["email"].(string)

			// var (
			// 	in  *nats.Msg
			// 	out *nats.Msg
			// 	err error
			// )

			// in = &nats.Msg{Subject: "profile.exists", Data: []byte(email)}
			// out, err = s.e.Conn().RequestMsg(in, 5*time.Second)
			// if err != nil {
			// 	s.m.Respond(w, r, err, http.StatusInternalServerError)
			// 	return
			// }

			// // pointer to struct
			// var p *struct {
			// 	ID uuid.UUID
			// 	// use a defined type
			// 	Email string
			// 	// use a defined type
			// 	Username string
			// 	// use a defined type
			// 	PhotoURL string
			// }

			// if err := events.Decode(out.Data, p); err != nil {
			// 	s.m.Respond(w, r, err, http.StatusInternalServerError)
			// 	return
			// }
		}
		// create a session (using JWT)

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
	s.e.Subscribe(events.EventTokenGenerateSignUp, s.handleConfirmSignUp())
}

func (s *Service) handleConfirmSignUp() nats.MsgHandler {
	// NOTE
	newReplyMsg := func(subj string, email, username string) (*nats.Msg, error) {
		encTk, err := s.tk.GenerateToken(context.Background(), jot.WithEnd(5*time.Minute), jot.WithPrivateClaims(jot.PrivateClaims{"email": email, "username": username}))
		if err != nil {
			return nil, err
		}

		// reply
		tokenGen := struct {
			Token []byte
		}{Token: encTk}
		p, err := events.Encode(tokenGen)
		if err != nil {
			return nil, err
		}

		return &nats.Msg{Subject: subj, Data: p}, nil
	}

	return func(msg *nats.Msg) {
		var tokenBd struct {
			Email    string
			Username string
		}

		if err := events.Decode(msg.Data, &tokenBd); err != nil {
			log.Println(err)
			return
		}

		msg, err := newReplyMsg(msg.Reply, tokenBd.Email, tokenBd.Username)
		if err != nil {
			log.Println(err)
			return
		}

		if err := s.e.Conn().PublishMsg(msg); err != nil {
			log.Println(err)
			return
		}

		log.Println("token generated")
	}
}
