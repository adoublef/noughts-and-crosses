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

	"github.com/hyphengolang/noughts-and-crosses/pkg/parse"
)

type Service struct {
	m  service.Router
	e  events.Broker
	tk jot.TokenClient
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.m.ServeHTTP(w, r)
}

func New(e events.Broker, tk jot.TokenClient) *Service {
	s := &Service{
		m:  service.NewRouter(),
		e:  e,
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
		Provider string `json:"provider"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var q request
		if err := s.m.Decode(w, r, &q); err != nil {
			s.m.Respond(w, r, err, http.StatusBadRequest)
			return
		}

		tk, err := s.tk.GenerateToken(context.Background(), jot.WithEnd(5*time.Minute), jot.WithPrivateClaims(jot.PrivateClaims{"email": q.Email}))
		if err != nil {
			s.m.Respond(w, r, err, http.StatusInternalServerError)
			return
		}

		msg, err := events.EncodeLoginConfirmMsg(q.Email, tk)
		if err != nil {
			s.m.Respond(w, r, err, http.StatusInternalServerError)
			return
		}

		if err := s.e.Publish(msg); err != nil {
			s.m.Respond(w, r, err, http.StatusInternalServerError)
			return
		}

		s.m.Respond(w, r, response{
			Provider: parse.ParseDomain(q.Email),
		}, http.StatusOK)
	}
}

func (s *Service) handleConfirmLogin() http.HandlerFunc {
	// type response struct{
	// 	OK bool `json:"ok"`
	// }
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := s.tk.ParseRequest(r)
		if err != nil {
			s.m.Respond(w, r, err, http.StatusUnauthorized)
			return
		}

		// ask registry service if user exists

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
		// create a session (using JWT)

		s.m.Respond(w, r, true, http.StatusOK)
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
	s.e.Subscribe(events.EventGenerateSignupToken, s.handleGenerateSignUpToken())
	s.e.Subscribe(events.EventVerifySignupToken, s.handleVerifySignUpToken())
}

func (s *Service) handleVerifySignUpToken() nats.MsgHandler {
	type A struct {
		Email string
	}

	return func(msg *nats.Msg) {
		var tokVal events.DataEmailToken
		if err := events.Decode(msg.Data, &tokVal); err != nil {
			log.Println(err)
			return
		}

		tk, err := s.tk.ParseToken(tokVal.Token)
		if err != nil {
			log.Println(err)
			return
		}

		email, _ := tk.PrivateClaims()["email"].(string)
		{
			raw := A{
				Email: email,
			}
			p, err := events.Encode(raw)
			if err != nil {
				log.Println(err)
				return
			}

			if err := msg.Respond(p); err != nil {
				log.Println(err)
				return
			}
		}

		log.Println("email", email)
	}
}

func (s *Service) handleGenerateSignUpToken() nats.MsgHandler {
	// NOTE move to events package when I can thing of a decent abstraction
	newReplyMsg := func(subj string, email string) (*nats.Msg, error) {
		encTk, err := s.tk.GenerateToken(context.Background(), jot.WithEnd(5*time.Minute), jot.WithPrivateClaims(jot.PrivateClaims{"email": email}))
		if err != nil {
			return nil, err
		}

		// reply
		tokenGen := events.DataEmailToken{Token: encTk}
		p, err := events.Encode(tokenGen)
		if err != nil {
			return nil, err
		}

		return &nats.Msg{Subject: subj, Data: p}, nil
	}

	return func(msg *nats.Msg) {
		var tokenBd events.DataSignUpConfirm
		if err := events.Decode(msg.Data, &tokenBd); err != nil {
			log.Println(err)
			return
		}

		msg, err := newReplyMsg(msg.Reply, tokenBd.Email)
		if err != nil {
			log.Println(err)
			return
		}

		if err := s.e.Publish(msg); err != nil {
			log.Println(err)
			return
		}

		log.Println("token generated")
	}
}
