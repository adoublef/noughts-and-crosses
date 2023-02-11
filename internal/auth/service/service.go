// https://cs.union.edu/~striegnk/courses/nlp-with-prolog/html/node93.html#:~:text=In%20parsing%2C%20we%20have%20a,correspond%20to%20the%20semantic%20representation.
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
	type Q struct {
		Email string `json:"email"`
	}

	type P struct {
		Provider string `json:"provider"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var q Q
		if err := s.m.Decode(w, r, &q); err != nil {
			s.m.Respond(w, r, err, http.StatusBadRequest)
			return
		}

		tk, err := s.t.SignToken(r.Context(), token.WithEnd(5*time.Minute), token.WithClaims(token.PrivateClaims{"email": q.Email}))
		if err != nil {
			s.m.Respond(w, r, err, http.StatusInternalServerError)
			return
		}

		if err := s.e.Conn().Publish(events.EventSendLoginConfirm, events.DataLoginConfirm{Email: q.Email, Token: tk}); err != nil {
			s.m.Respond(w, r, err, http.StatusInternalServerError)
			return
		}

		s.m.Respond(w, r, P{
			Provider: parse.ParseDomain(q.Email),
		}, http.StatusOK)
	}
}

func (s *Service) handleConfirmLogin() http.HandlerFunc {
	type P struct {
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

		s.m.Respond(w, r, P{
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
	s.e.Conn().Subscribe(events.EventGenerateSignupToken, s.generateSignupToken())
	// responds back to `registry`
	s.e.Conn().Subscribe(events.EventVerifySignupToken, s.verifySignupToken())
	// responds back to `registry`
	s.e.Conn().Subscribe(events.EventCreateProfileValidation, s.verifyCreateProfileToken())
}

func (s *Service) verifyCreateProfileToken() nats.MsgHandler {
	type D struct{ events.Data[struct{}] }

	return func(msg *nats.Msg) {
		var d D

		var payload events.DataAuthToken
		if err := events.Unmarshal(msg.Data, &payload); err != nil {
			msg.Respond(d.Errorf("failed to decode data: %v", err))
			return
		}

		tk, err := s.t.ParseToken(payload.Token)
		if err != nil {
			msg.Respond(d.Errorf("failed to parse token: %v", err))
			return
		}

		if email := tk.PrivateClaims()["email"]; email != payload.Email {
			msg.Respond(d.Errorf("something went wrong with the verifying identity"))

			return
		}

		// emails match so this is ok!
		if err := msg.Respond(d.Bytes()); err != nil {
			log.Printf("messages response: %v", err)
		}

	}
}

func (s *Service) verifySignupToken() nats.MsgHandler {
	type D struct{ events.Data[string] }

	return func(msg *nats.Msg) {
		var d D

		var payload events.DataToken
		if err := s.e.Conn().Enc.Decode(msg.Subject, msg.Data, &payload); err != nil {
			msg.Respond(d.Errorf("decode msg: %v", err))
			return
		}

		jwt, err := s.t.ParseToken(payload.Token)
		if err != nil {
			msg.Respond(d.Errorf("parse token: %v", err))
			return
		}

		d.Value = jwt.PrivateClaims()["email"].(string)
		if err := msg.Respond(d.Bytes()); err != nil {
			log.Printf("messages response: %v", err)
		}

	}
}

func (s *Service) generateSignupToken() nats.MsgHandler {
	type D struct{ events.Data[[]byte] }

	return func(msg *nats.Msg) {
		var d D

		var payload events.DataEmail
		if err := events.Unmarshal(msg.Data, &payload); err != nil {
			// log.Println(err)
			msg.Respond(d.Errorf("decode error %v", err))
			return
		}

		tk, err := s.t.SignToken(context.Background(), token.WithEnd(5*time.Minute), token.WithClaims(token.PrivateClaims{"email": payload.Email}))
		if err != nil {
			// log.Println(err)
			msg.Respond(d.Errorf("sign token: %v", err))

			return
		}

		// set value
		d.Value = tk
		if err := msg.Respond(d.Bytes()); err != nil {
			log.Printf("messages response: %v", err)
		}
	}
}
