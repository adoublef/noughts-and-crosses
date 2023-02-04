package service

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/hyphengolang/noughts-and-crosses/internal/events"
	"github.com/hyphengolang/noughts-and-crosses/internal/reg"
	repo "github.com/hyphengolang/noughts-and-crosses/internal/reg/repository"
	"github.com/hyphengolang/noughts-and-crosses/internal/service"
	"github.com/jackc/pgx/v5"
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

func New(nc *nats.Conn) *Service {
	s := &Service{
		m: service.NewRouter(),
		e: events.NewClient(nc),
	}
	s.routes()
	return s
}

func (s *Service) routes() {
	s.m.Post("/users", s.handleRegistration())
	s.m.Get("/users/verify", s.handleVerifyRegistration())

	r := s.m.With(service.PathParam("uuid", uuidParser))
	r.Delete("/users/{uuid}", s.handleTermination())
	r.Get("/users/{uuid}/profile", s.handleGetProfile())
	r.Post("/users/{uuid}/profile", s.handleSetProfile())
	r.Put("/users/{uuid}/profile/photo-url", s.handleSetPhotoURL())
}

func (s *Service) handleTermination() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := uuidFromRequest(r)

		args := pgx.NamedArgs{"id": userID}
		if err := s.r.UnsetProfile(r.Context(), args); err != nil {
			s.m.Respond(w, r, err, http.StatusInternalServerError)
			return
		}

		s.m.Respond(w, r, userID, http.StatusOK)
	}
}

func (s *Service) handleGetProfile() http.HandlerFunc {
	type P struct {
		Profile *reg.User `json:"profile"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		uid, _ := uuidFromRequest(r)

		args := pgx.NamedArgs{"id": uid}
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

		args := pgx.NamedArgs{"id": id, "photo_url": q.PhotoURL}
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

		args := pgx.NamedArgs{
			"id":       uid,
			"username": q.Username,
			"bio":      q.Bio,
		}
		if err := s.r.UpdateProfile(r.Context(), args); err != nil {
			s.m.Respond(w, r, err, http.StatusInternalServerError)
			return
		}

		// profile created. add location header
		s.m.Respond(w, r, nil, http.StatusOK)
	}
}

func (s *Service) handleRegistration() http.HandlerFunc {
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

		args := pgx.NamedArgs{
			"id":       uuid.New(),
			"email":    q.Email,
			"username": q.Username,
		}

		if err := s.r.SetProfile(r.Context(), args); err != nil {
			s.m.Respond(w, r, err, http.StatusInternalServerError)
			return
		}

		// NOTE send confirmation email with verification link
		if _, err := s.e.Request("user.registered", q, 5*time.Second); err != nil {
			s.m.Respond(w, r, err, http.StatusInternalServerError)
			return
		}

		p := &P{
			Message: "Check emails to continue registration process",
		}

		s.m.Respond(w, r, p, http.StatusOK)
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
