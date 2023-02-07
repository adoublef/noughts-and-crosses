package service

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/hyphengolang/noughts-and-crosses/internal/events"
	"github.com/hyphengolang/noughts-and-crosses/internal/reg"
	repo "github.com/hyphengolang/noughts-and-crosses/internal/reg/repository"
	"github.com/hyphengolang/noughts-and-crosses/internal/service"
	"github.com/nats-io/nats.go"
)

var _ http.Handler = (*Service)(nil)

type Service struct {
	m service.Router
	e events.Broker
	r repo.Repo
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.m.ServeHTTP(w, r)
}

// events.Client should be a dependency
func New(e events.Broker, r repo.Repo) *Service {
	s := &Service{
		m: service.NewRouter(),
		e: e,
		r: r,
	}
	go s.listen()
	s.routes()
	return s
}

func (s *Service) routes() {
	s.m.Post("/signup", s.handleSignUp())
	s.m.Get("/signup", s.handleConfirmSignUp())

	s.m.Post("/users", s.handleRegisterProfile())
	s.m.Get("/users/verify", s.handleVerifyRegistration())

	// r := s.m.With(service.PathParam("uuid", uuidParser))
	// r.Delete("/users/{uuid}", s.handleTermination())
	// r.Get("/users/{uuid}/profile", s.handleGetProfile())
	// r.Post("/users/{uuid}/profile", s.handleSetProfile())
	// r.Put("/users/{uuid}/profile/photo-url", s.handleSetPhotoURL())
}

func (s *Service) handleConfirmSignUp() http.HandlerFunc {
	parseHeader := func(r *http.Request) (string, error) {
		token := strings.TrimSpace(r.URL.Query().Get("token"))
		if token == "" {
			return "", fmt.Errorf(`empty query (Authorization)`)
		}
		return strings.TrimSpace(strings.TrimPrefix(token, "Bearer")), nil
	}

	newSignupTokenMsg := func(token string) (*nats.Msg, error) {
		v := struct {
			Token string
		}{Token: token}
		p, err := events.Encode(v)
		if err != nil {
			return nil, err
		}
		// Request from Auth service to get token from header.
		msg := nats.Msg{
			Subject: events.EventUserSignup,
			// token from header
			Data: p,
		}
		return &msg, nil
	}

	return func(w http.ResponseWriter, r *http.Request) {
		token, err := parseHeader(r)
		if err != nil {
			s.m.Respond(w, r, err, http.StatusBadRequest)
			return
		}

		msg, err := newSignupTokenMsg(token)
		if err != nil {
			s.m.Respond(w, r, err, http.StatusInternalServerError)
			return
		}

		raw, err := s.e.Conn().RequestMsg(msg, 5*time.Second)
		if err != nil {
			s.m.Respond(w, r, err, http.StatusBadRequest)
			return
		}

		_ = raw

		// Decode token from raw out.
		// var out struct {
		// }
		// if err := events.Decode(raw.Data, &out); err != nil {
		// 	s.m.Respond(w, r, err, http.StatusInternalServerError)
		// 	return
		// }

		s.m.Respond(w, r, nil, http.StatusOK)
	}
}

func (s *Service) handleSignUp() http.HandlerFunc {
	type Q struct {
		Email    string `json:"email"`
		Username string `json:"username"`
	}

	type P struct {
		Message string `json:"message"`
	}

	newSignUpMsg := func(email, username string) (*nats.Msg, error) {
		// send email to complete sign-up process
		// automatically check which email provider so
		// can send a link to the correct email provider
		// https://www.freecodecamp.org/news/the-best-free-email-providers-2021-guide-to-online-email-account-services/
		data := struct {
			Email    string
			Username string
		}{Email: email, Username: username}
		p, err := events.Encode(data)
		if err != nil {
			return nil, err
		}

		return &nats.Msg{Subject: events.EventUserSignup, Data: p}, nil
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var q Q
		if err := s.m.Decode(w, r, &q); err != nil {
			s.m.Respond(w, r, err, http.StatusBadRequest)
			return
		}

		msg, err := newSignUpMsg(q.Email, q.Username)
		if err != nil {
			s.m.Respond(w, r, err, http.StatusInternalServerError)
			return
		}

		if err := s.e.Conn().PublishMsg(msg); err != nil {
			s.m.Respond(w, r, err, http.StatusInternalServerError)
			return
		}

		s.m.Respond(w, r, P{Message: "email sent"}, http.StatusAccepted)
	}
}

func (s *Service) handleTermination() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, _ := uuidFromRequest(r)

		args := repo.UUIDArgs{ID: uid}
		if err := s.r.UnsetProfile(r.Context(), args); err != nil {
			s.m.Respond(w, r, err, http.StatusInternalServerError)
			return
		}

		s.m.Respond(w, r, uid, http.StatusOK)
	}
}

func (s *Service) handleGetProfile() http.HandlerFunc {
	type P struct {
		Profile *reg.Profile `json:"profile"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		uid, _ := uuidFromRequest(r)

		args := repo.UUIDArgs{ID: uid}
		profile, err := s.r.GetProfile(r.Context(), args)
		if err != nil {
			s.m.Respond(w, r, err, http.StatusInternalServerError)
			return
		}

		p := P{Profile: profile}
		s.m.Respond(w, r, p, http.StatusOK)
	}
}

// TODO experimental
func (s *Service) handleSetPhotoURL() http.HandlerFunc {
	type Q struct {
		PhotoURL string `json:"photoUrl"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		id, _ := uuidFromRequest(r)

		var q Q
		if err := s.m.Decode(w, r, &q); err != nil {
			s.m.Respond(w, r, err, http.StatusBadRequest)
			return
		}

		args := repo.SetPhotoURLArgs{
			ID:       id,
			PhotoURL: q.PhotoURL,
		}

		if err := s.r.SetPhotoURL(r.Context(), args); err != nil {
			s.m.Respond(w, r, err, http.StatusInternalServerError)
			return
		}

		// profile created. add location header
		s.m.Respond(w, r, nil, http.StatusOK)
	}
}

func (s *Service) handleSetProfile() http.HandlerFunc {
	type Q struct {
		Username string `json:"username"`
		Bio      string `json:"bio"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		uid, _ := uuidFromRequest(r)

		var q Q
		if err := s.m.Decode(w, r, &q); err != nil {
			s.m.Respond(w, r, err, http.StatusBadRequest)
			return
		}

		args := repo.UpdateProfileArgs{
			ID:       uid,
			Username: q.Username,
			Bio:      q.Bio,
		}

		if err := s.r.UpdateProfile(r.Context(), args); err != nil {
			s.m.Respond(w, r, err, http.StatusInternalServerError)
			return
		}

		// profile created. add location header
		s.m.Respond(w, r, nil, http.StatusOK)
	}
}

func (s *Service) handleRegisterProfile() http.HandlerFunc {
	type Q struct {
		Email    string `json:"email"`
		Username string `json:"username"`
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

		args := repo.SetProfileArgs{
			Email:    q.Email,
			Username: q.Username,
		}

		if err := s.r.SetProfile(r.Context(), args); err != nil {
			s.m.Respond(w, r, err, http.StatusInternalServerError)
			return
		}

		// NOTE send confirmation email with verification link
		// if err := s.e.Publish(events.EventUserRegister, q); err != nil {
		// 	s.m.Respond(w, r, err, http.StatusInternalServerError)
		// 	return
		// }

		p := &P{
			Message: "Check emails to continue registration process",
		}

		s.m.Respond(w, r, p, http.StatusCreated)
	}
}

func (s *Service) handleVerifyRegistration() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get token from url param and verify with the token stored in database
	}
}

func uuidParser(r *http.Request, key string) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, key))
}

func uuidFromRequest(r *http.Request) (uuid.UUID, error) {
	return service.PathParamFromRequest[uuid.UUID](r)
}

//  Events

func (s *Service) listen() {
	// listen to events
}
