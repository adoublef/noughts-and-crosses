package service

import (
	"net/http"
	"redis-ws/pkg/websocket"

	"github.com/hyphengolang/noughts-and-crosses/internal/service"
	"github.com/hyphengolang/noughts-and-crosses/pkg/websocket/pubsub"
)

type Option func(*Service)

type Service struct {
	m service.Router
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.m.ServeHTTP(w, r)
}

func New(opts ...Option) *Service {
	s := &Service{
		m: service.NewRouter(),
	}

	for _, opt := range opts {
		opt(s)
	}

	s.routes()
	return s
}

func (s *Service) routes() {
	s.m.Get("/", func(w http.ResponseWriter, r *http.Request) {
		s.m.Respond(w, r, "Hello, world!", http.StatusOK)
	})

	// websocket endpoint
	s.m.Get("/play", websocket.NewClient(2, pubsub.New(256)).ServeHTTP)
}
