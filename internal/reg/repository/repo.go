package repo

import (
	"context"

	"github.com/google/uuid"
	pg "github.com/hyphengolang/smtp.google/internal/postgres"
	"github.com/hyphengolang/smtp.google/internal/reg"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID       uuid.UUID
	Email    string
	Username string
	// PhotoURL is a link to the GCloud Bucket
	PhotoURL *string //Optional
	Verified bool    //default=false
}

type Repo interface {
	SetProfile(ctx context.Context, args pgx.QueryRewriter) error
	GetProfile(ctx context.Context, args pgx.QueryRewriter) (*reg.User, error)
	UnsetProfile(ctx context.Context, args pgx.QueryRewriter) error
	UpdateProfile(ctx context.Context, args pgx.QueryRewriter) error
	SetPhotoURL(ctx context.Context, args pgx.QueryRewriter) error
}

var _ Repo = (*repo)(nil)

type repo struct {
	c pg.Conn[reg.User]
}

// GetProfile implements Repo
func (r *repo) GetProfile(ctx context.Context, args pgx.QueryRewriter) (*reg.User, error) {
	const q = `
	SELECT id, email, username
	FROM reg.user
	WHERE id = @id`

	return r.c.QueryRowContext(ctx, func(r pgx.Row, u *reg.User) error {
		return r.Scan(&u.ID, &u.Email, &u.Username)
	}, q, args)
}

func (r *repo) SetProfile(ctx context.Context, args pgx.QueryRewriter) error {
	const q = `
	INSERT INTO reg.user (id, email, username)
	VALUES (@id, @email, @username)`

	_, err := r.c.ExecContext(ctx, q, args)
	return err
}

func (r *repo) UnsetProfile(ctx context.Context, args pgx.QueryRewriter) error {
	const q = `
	DELETE FROM reg.user
	WHERE id = @id`

	count, err := r.c.ExecContext(ctx, q, args)
	if count == 0 {
		return pg.ErrNoRowsAffected
	}

	return err
}

func (r *repo) SetBio(ctx context.Context, args pgx.QueryRewriter) error {
	const q = `
	UPDATE reg.user
	SET bio = NULLIF(@bio,'')
	WHERE id = @id`

	count, err := r.c.ExecContext(ctx, q, args)
	if count == 0 {
		return pg.ErrNoRowsAffected
	}

	return err
}

func (r *repo) SetPhotoURL(ctx context.Context, args pgx.QueryRewriter) error {
	const q = `
	UPDATE reg.user
	SET photo_url = NULLIF(@photo_url,'')
	WHERE id = @id`

	count, err := r.c.ExecContext(ctx, q, args)
	if count == 0 {
		return pg.ErrNoRowsAffected
	}

	return err
}

func (r *repo) UpdateProfile(ctx context.Context, args pgx.QueryRewriter) error {
	const q = `
	UPDATE reg.user
	SET username = @username,
		bio = NULLIF(@bio,''),
		WHERE id = @id`

	count, err := r.c.ExecContext(ctx, q, args)
	if count == 0 {
		return pg.ErrNoRowsAffected
	}

	return err
}

func New(rwc *pgxpool.Pool) Repo {
	r := &repo{c: pg.NewConn[reg.User](rwc)}
	return r
}
