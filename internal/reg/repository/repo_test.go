package repo_test

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hyphengolang/prelude/testing/is"
	"github.com/hyphengolang/smtp.google/internal/docker"
	repo "github.com/hyphengolang/smtp.google/internal/reg/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	regRepo   repo.Repo
	container *docker.PostgresContainer
)

func init() {
	ctx := context.TODO()

	m := `
	CREATE SCHEMA reg;

	CREATE EXTENSION IF NOT EXISTS pgcrypto;
	CREATE EXTENSION IF NOT EXISTS citext;

	CREATE TABLE IF NOT EXISTS reg.user (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		email CITEXT UNIQUE NOT NULL CHECK (email ~ '^[a-zA-Z0-9.!#$%&’*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,253}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,253}[a-zA-Z0-9])?)*$'),
		username VARCHAR(15) UNIQUE NOT NULL CHECK (username <> ''),
		bio VARCHAR(160),
		photo_url TEXT
	);
	`

	var (
		conn *pgxpool.Pool
		err  error
	)

	container, conn, err = docker.NewPostgresConnection(ctx, "5432/tcp", 15*time.Second, m)
	if err != nil {
		log.Fatal(err)
	}

	// initialize test repo
	regRepo = repo.New(conn)
}

func TestUserRepository(t *testing.T) {
	is := is.New(t)
	ctx := context.Background()

	johnDoe := uuid.New()

	t.Run("create a new row for user", func(t *testing.T) {
		args := pgx.NamedArgs{
			"id":       johnDoe,
			"email":    "john@doe.com",
			"username": "john123doe",
		}

		err := regRepo.SetProfile(ctx, args)
		is.NoErr(err) // create a new profile
	})

	t.Run("update photo_url for created user", func(t *testing.T) {
		args := pgx.NamedArgs{
			"id":        johnDoe,
			"photo_url": "https://link/to/bucket.com/someId",
		}

		err := regRepo.SetPhotoURL(ctx, args)
		is.NoErr(err) // update profile photo
	})

	t.Run("update photo_url fail", func(t *testing.T) {
		args := pgx.NamedArgs{
			"id":        uuid.New(),
			"photo_url": "https://link/to/bucket.com/someId",
		}

		err := regRepo.SetPhotoURL(ctx, args)
		is.True(err != nil) // failed to update profile photo
	})

	t.Run("get profile for 'john doe'", func(t *testing.T) {
		args := pgx.NamedArgs{
			"id": johnDoe,
		}

		user, err := regRepo.GetProfile(ctx, args)
		is.NoErr(err)                        // get profile
		is.Equal(user.ID, johnDoe)           // id is correct
		is.Equal(user.Email, "john@doe.com") // email is correct
	})

	t.Run("delete profile for 'john doe'", func(t *testing.T) {
		args := pgx.NamedArgs{
			"id": johnDoe,
		}

		err := regRepo.UnsetProfile(ctx, args)
		is.NoErr(err) // delete profile
	})
}
