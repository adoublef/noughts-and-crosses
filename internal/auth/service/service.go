package service

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/hyphengolang/noughts-and-crosses/internal/events"
	"github.com/hyphengolang/noughts-and-crosses/internal/service"
	token "github.com/hyphengolang/noughts-and-crosses/pkg/auth/jwt"

	"github.com/hyphengolang/noughts-and-crosses/pkg/parse"
)

type Service struct {
	m service.Router
	e events.Broker
	t token.Client
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.m.ServeHTTP(w, r)
}

func New(e events.Broker, t token.Client) *Service {
	s := &Service{
		m: service.NewRouter(),
		e: e,
		t: t,
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

		tk, err := s.t.GenerateToken(context.Background(), token.WithEnd(5*time.Minute), token.WithPrivateClaims(token.PrivateClaims{"email": q.Email}))
		if err != nil {
			s.m.Respond(w, r, err, http.StatusInternalServerError)
			return
		}

		msg, err := events.NewSendLoginConfirmMsg(q.Email, tk)
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
	type response struct {
		Username     string  `json:"username"`
		AccessToken  string  `json:"accessToken"`
		RefreshToken string  `json:"refreshToken"`
		PhotoURL     *string `json:"photoUrl"` //optional
	}

	return func(w http.ResponseWriter, r *http.Request) {
		tk, err := s.t.ParseRequest(r)
		if err != nil {
			s.m.Respond(w, r, err, http.StatusUnauthorized)
			return
		}

		// TODO: ASK REGISTRY IF USER EXISTS OR NOT
		email, _ := tk.PrivateClaims()["email"].(string)

		// TODO define a structure for this package
		// raw, err := s.e.Request(&nats.Msg{Data: []byte(email)}, 5*time.Second)
		// if err != nil {
		// 	s.m.Respond(w, r, err, http.StatusNotFound)
		// 	return
		// }

		// should include username, photoUrl & uid to be added to a jwt.
		// data := struct {
		// 	ID       uuid.UUID
		// 	Username string
		// 	PhotoURL *string //optional
		// }{}
		// if err := events.Decode(raw, &data); err != nil {
		// 	s.m.Respond(w, r, err, http.StatusInternalServerError)
		// 	return
		// }

		// claims := jot.PrivateClaims{}
		// claims := jot.PrivateClaims{
		// 	"username": data.Username,
		// 	"photoUrl": data.PhotoURL,
		// }
		// accessToken, _ := s.tk.GenerateToken(r.Context(), jot.WithEnd(30*time.Minute), jot.WithPrivateClaims(claims), jot.WithSubject(data.ID.String()))
		// if err != nil {
		// 	s.m.Respond(w, r, err, http.StatusInternalServerError)
		// 	return
		// }
		// claims["email"] = email
		// refreshToken, _ := s.tk.GenerateToken(r.Context(), jot.WithEnd(7*24*time.Hour), jot.WithPrivateClaims(claims), jot.WithSubject(data.ID.String()))
		// if err != nil {
		// 	s.m.Respond(w, r, err, http.StatusInternalServerError)
		// 	return
		// }

		s.m.Respond(w, r, response{
			// AccessToken:  string(accessToken),
			// RefreshToken: string(refreshToken),
			Username: "tmp:" + email,
			// Username:     data.Username,
			// PhotoURL:     data.PhotoURL,
		}, http.StatusOK)
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
	// responds back to `mailing`
	s.e.Subscribe(events.EventGenerateSignupToken, s.generateSignupToken())
	// responds back to `registry`
	s.e.Subscribe(events.EventVerifySignupToken, s.verifySignupToken())
	// responds back to `registry`
	s.e.Subscribe(events.EventCreateProfileValidation, s.verifyCreateProfileToken())
}

func (s *Service) verifyCreateProfileToken() nats.MsgHandler {
	type Data struct{ events.Data[struct{}] }

	return func(msg *nats.Msg) {
		var result Data

		var data events.DataAuthToken
		if err := events.Unmarshal(msg.Data, &data); err != nil {
			p := result.Errorf("failed to decode data: %v", err)
			msg.Respond(p)
			return
		}

		tk, err := s.t.ParseToken(data.Token)
		if err != nil {
			p := result.Errorf("failed to parse token: %v", err)
			msg.Respond(p)
			return
		}

		if email := tk.PrivateClaims()["email"]; email != data.Email {
			p := result.Errorf("something went wrong with the verifying identity")
			msg.Respond(p)
			return
		}

		// emails match so this is ok!
		msg.Respond(result.Bytes())
	}
}

func (s *Service) verifySignupToken() nats.MsgHandler {
	type Data struct {
		events.Data[string]
	}

	return func(msg *nats.Msg) {
		var reply Data

		var in events.DataEmailToken
		if err := events.Unmarshal(msg.Data, &in); err != nil {
			msg.Respond(reply.Errorf("decode msg: %v", err))
			return
		}

		tk, err := s.t.ParseToken(in.Token)
		if err != nil {
			msg.Respond(reply.Errorf("parse token: %v", err))
			return
		}

		// Could be work checking
		email := tk.PrivateClaims()["email"].(string)
		// {
		reply.Value = email
		msg.Respond(reply.Bytes())
		// raw := A{
		// 	Email: email,
		// }
		// p, err := events.Encode(raw)
		// if err != nil {
		// 	log.Println(err)
		// 	return
		// }

		// if err := msg.Respond(p); err != nil {
		// 	log.Println(err)
		// 	return
		// }
		// }
		log.Println("email", email)
	}
}

func (s *Service) generateSignupToken() nats.MsgHandler {
	type Data struct{ events.Data[[]byte] }

	return func(msg *nats.Msg) {
		var reply Data

		var in events.DataSignUpConfirm
		if err := events.Unmarshal(msg.Data, &in); err != nil {
			log.Println(err)
			msg.Respond(reply.Errorf("decode error %v", err))
			return
		}

		token, err := s.t.GenerateToken(context.Background(), token.WithEnd(5*time.Minute), token.WithPrivateClaims(token.PrivateClaims{"email": in.Email}))
		if err != nil {
			log.Println(err)
			msg.Respond(reply.Errorf("sign token: %v", err))
			return
		}

		// set value
		reply.Value = token
		if err := msg.Respond(reply.Bytes()); err != nil {
			log.Println(err)
		}
	}
}
