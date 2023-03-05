package service

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/hyphengolang/noughts-and-crosses/internal/events"
	"github.com/hyphengolang/noughts-and-crosses/internal/registry"
	repo "github.com/hyphengolang/noughts-and-crosses/internal/registry/repository"
	"github.com/hyphengolang/noughts-and-crosses/internal/service"
	"github.com/hyphengolang/noughts-and-crosses/pkg/parse"
)

func uuidParser(r *http.Request, key string) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, key))
}

func uuidFromRequest(r *http.Request) (uuid.UUID, error) {
	return service.PathParamFromRequest[uuid.UUID](r)
}

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
	// NOTE: may be more appropriate for this to hang on auth service
	s.m.Get("/signup", s.handleVerifySignup())

	s.m.Post("/users", s.handleRegisterProfile())

	// r := s.m.With(service.PathParam("uuid", uuidParser))
	// r.Delete("/users/{uuid}", s.handleTermination())
	// r.Get("/users/{uuid}/profile", s.handleGetProfile())
	// r.Post("/users/{uuid}/profile", s.handleSetProfile())
	// r.Put("/users/{uuid}/profile/photo-url", s.handleSetPhotoURL())
}

func (s *Service) handleVerifySignup() http.HandlerFunc {
	parseEmail := func(r *http.Request, token []byte, timeout time.Duration) (email string, err error) {
		var reply struct{ events.Data[string] }
		err = s.e.Conn().Request(events.EventVerifySignupToken, events.DataToken{Token: token}, &reply, 5*time.Second)
		if err != nil {
			return
		}

		return reply.Value, reply.Err
	}

	type P struct {
		Email string `json:"email"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := parse.ParseToken(r)

		if err != nil {
			// log.Printf("parsing header")
			s.m.Respond(w, r, err, http.StatusUnauthorized)
			return
		}

		email, err := parseEmail(r, token, 5*time.Second)

		if err != nil {
			// log.Printf("parsing token")
			s.m.Respond(w, r, err, http.StatusUnauthorized)
			return
		}

		p := P{
			Email: email,
		}
		s.m.Respond(w, r, p, http.StatusOK)
	}
}

func (s *Service) handleSignUp() http.HandlerFunc {
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

		if err := s.e.Conn().Publish(events.EventSendSignupConfirm, events.DataEmail{Email: q.Email}); err != nil {
			s.m.Respond(w, r, err, http.StatusInternalServerError)
			return
		}

		s.m.Respond(w, r, P{
			Provider: parse.ParseDomain(q.Email),
			// VerificationToken: token,
		}, http.StatusAccepted)
	}
}

// should appear when sign-up confirm email returns authorized
// the token has the email & username so that should be sent
// here along with user bio (optional) and user image (also optional)
// then on confirm, it will create the user profile
// and redirect to dashboard.
// Application already verified the email so no need to auth again.
// NOTE: this means signup confirm only needs email address
// add username can be provided at this endpoint. decluttering the API a bit
func (s *Service) handleRegisterProfile() http.HandlerFunc {
	type Q struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Bio      string `json:"bio"`
	}

	auth := func(w http.ResponseWriter, r *http.Request, email string) error {
		type D struct{ events.Data[struct{}] }

		token, err := parse.ParseToken(r)
		if err != nil {
			return err
		}

		var data D
		err = s.e.Conn().Request(events.EventCreateProfileValidation, events.DataAuthToken{Token: token, Email: email}, &data, 5*time.Second)
		if err != nil {
			return err
		}

		return data.Err

		// msg, err := events.NewCreateProfileValidationMsg(email, token)
		// if err != nil {
		// 	return err
		// }

		// // this will be an a gob encoded message so needs to be decoded
		// p, err := s.e.Request(msg, 5*time.Second)
		// if err != nil {
		// 	return err
		// }

		// var auth Data
		// if err := events.Unmarshal(p, &auth); err != nil {
		// 	return err
		// }

		// return auth.Err
	}

	type P struct {
		Username   string `json:"username"`
		ProfileURL string `json:"profileUrl"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var q Q
		if err := s.m.Decode(w, r, &q); err != nil {
			s.m.Respond(w, r, err, http.StatusBadRequest)
			return
		}

		if err := auth(w, r, q.Email); err != nil {
			s.m.Respond(w, r, err, http.StatusUnauthorized)
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
		s.m.Respond(w, r, P{
			Username:   q.Username,
			ProfileURL: s.m.ClientURI() + "/todo",
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

//  Events

func (s *Service) listen() {
	// listen to events
}
