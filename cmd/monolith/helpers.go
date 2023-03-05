package main

import (
	"context"
	"log"

	authHTTP "github.com/hyphengolang/noughts-and-crosses/internal/auth/service"
	"github.com/hyphengolang/noughts-and-crosses/internal/conf"
	"github.com/hyphengolang/noughts-and-crosses/internal/events"
	gamesHTTP "github.com/hyphengolang/noughts-and-crosses/internal/games/service"
	mailHTTP "github.com/hyphengolang/noughts-and-crosses/internal/mailing/service"
	registryDB "github.com/hyphengolang/noughts-and-crosses/internal/registry/repository"
	registryHTTP "github.com/hyphengolang/noughts-and-crosses/internal/registry/service"
	"github.com/hyphengolang/noughts-and-crosses/internal/smtp"
	token "github.com/hyphengolang/noughts-and-crosses/pkg/auth/jwt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
)

func newMailingService(nc *nats.EncodedConn) *mailHTTP.Service {
	em := smtp.NewMailer(conf.SMTPUsername, conf.SMTPPassword, conf.SMTPHost, 587)
	ec := events.NewClient(nc)
	return mailHTTP.New(em, ec)
}

func newRegService(nc *nats.EncodedConn, pg *pgxpool.Pool) *registryHTTP.Service {
	ec := events.NewClient(nc)
	return registryHTTP.New(ec, registryDB.New(pg))
}

func newAuthService(nc *nats.EncodedConn) *authHTTP.Service {
	tk := token.NewTokenClient(token.WithPEM(conf.JWTSecret))
	ec := events.NewClient(nc)
	return authHTTP.New(ec, tk)
}

func newGamesService() *gamesHTTP.Service {
	return gamesHTTP.New()
}

func newNATService() (*nats.EncodedConn, error) {
	nc, err := nats.Connect(conf.NATSURI, nats.UserJWTAndSeed(conf.NATSToken, conf.NATSSeed), nats.ErrorHandler(func(nc *nats.Conn, s *nats.Subscription, err error) {
		if s != nil {
			log.Printf("Async error in %q/%q: %v", s.Subject, s.Queue, err)
		} else {
			log.Printf("Async error outside subscription: %v", err)
		}
	}))
	if err != nil {
		return nil, err
	}

	return nats.NewEncodedConn(nc, nats.GOB_ENCODER)
}

func newDBConn(ctx context.Context) (*pgxpool.Pool, error) {
	conn, err := pgxpool.New(ctx, conf.DBURL)
	if err != nil {
		log.Fatal(err)
	}

	return conn, conn.Ping(ctx)
}
