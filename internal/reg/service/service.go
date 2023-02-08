package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/hyphengolang/noughts-and-crosses/internal/conf"
	"github.com/hyphengolang/noughts-and-crosses/internal/events"
	"github.com/hyphengolang/noughts-and-crosses/internal/reg"
	repo "github.com/hyphengolang/noughts-and-crosses/internal/reg/repository"
	"github.com/hyphengolang/noughts-and-crosses/internal/service"
	"github.com/hyphengolang/noughts-and-crosses/pkg/parse"
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
	type response struct {
		Email string `json:"email"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.String())

		token, err := parse.ParseHeader(r)
		if err != nil {
			s.m.Respond(w, r, err, http.StatusUnauthorized)
			return
		}

		p := response{}
		{
			msg, err := events.EncodeSignupVerifyMsg(token)
			if err != nil {
				s.m.Respond(w, r, err, http.StatusInternalServerError)
				return
			}

			// get the email from token, parsed from auth
			out, err := s.e.Request(msg, 5*time.Second)
			if err != nil {
				s.m.Respond(w, r, err, http.StatusInternalServerError)
				return
			}

			var raw struct {
				Email string
			}

			if err := events.Decode(out, &raw); err != nil {
				s.m.Respond(w, r, err, http.StatusInternalServerError)
				return
			}

			p.Email = raw.Email
		}

		fmt.Println(p)
		s.m.Respond(w, r, p, http.StatusOK)
	}
}

func (s *Service) handleSignUp() http.HandlerFunc {
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

		msg, err := events.EncodeSendSignupConfirmMsg(q.Email)
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
		}, http.StatusAccepted)
	}
}

func (s *Service) handleRegisterProfile() http.HandlerFunc {
	type request struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Bio      string `json:"bio"`
	}

	type response struct {
		Username string `json:"username"`
		Location string `json:"location"`
	}

	// should appear when sign-up confirm email returns authorized
	// the token has the email & username so that should be sent
	// here along with user bio (optional) and user image (also optional)
	// then on confirm, it will create the user profile
	// and redirect to dashboard.
	// Application already verified the email so no need to auth again.
	// NOTE: this means signup confirm only needs email address
	// add username can be provided at this endpoint. decluttering the API a bit
	return func(w http.ResponseWriter, r *http.Request) {
		// use the same token as signup confirm, this should be protected
		{
			_, err := parse.ParseHeader(r)
			if err != nil {
				s.m.Respond(w, r, err, http.StatusUnauthorized)
				return
			}

			// msg, err := events.EncodeSignupTokenMsg(token)
			// if err != nil {
			// 	s.m.Respond(w, r, err, http.StatusInternalServerError)
			// 	return
			// }

			// _, err = s.e.Request(msg, 5*time.Second)
			// if err != nil {
			// 	s.m.Respond(w, r, err, http.StatusBadRequest)
			// 	return
			// }
		}
		var q request
		if err := s.m.Decode(w, r, &q); err != nil {
			s.m.Respond(w, r, err, http.StatusBadRequest)
			return
		}

		args := repo.SetProfileArgs{
			Email:    q.Email,
			Username: q.Username,
			Bio:      q.Bio,
		}

		if err := s.r.SetProfile(r.Context(), args); err != nil {
			s.m.Respond(w, r, err, http.StatusInternalServerError)
			return
		}

		// TODO: send email to user with verification link
		// TODO: Update Location header
		s.m.Respond(w, r, response{
			Username: q.Username,
			Location: conf.ClientURI + "/todo",
		}, http.StatusCreated)
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
