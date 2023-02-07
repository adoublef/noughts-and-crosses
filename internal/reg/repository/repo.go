package repo

import (
	"context"

	"github.com/google/uuid"
	pg "github.com/hyphengolang/noughts-and-crosses/internal/postgres"
	"github.com/hyphengolang/noughts-and-crosses/internal/reg"
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
	GetProfile(ctx context.Context, args pgx.QueryRewriter) (*reg.Profile, error)
	UnsetProfile(ctx context.Context, args pgx.QueryRewriter) error
	UpdateProfile(ctx context.Context, args pgx.QueryRewriter) error
	SetPhotoURL(ctx context.Context, args pgx.QueryRewriter) error
}

type repo struct {
	c pg.Conn[reg.Profile]
}

type UUIDArgs struct {
	ID uuid.UUID
}

func (a UUIDArgs) RewriteQuery(ctx context.Context, conn *pgx.Conn, sql string, args []any) (newSQL string, newArgs []any, err error) {
	na := pgx.NamedArgs{
		"id": a.ID,
	}

	return na.RewriteQuery(ctx, conn, sql, args)
}

// GetProfile implements Repo
func (r *repo) GetProfile(ctx context.Context, args pgx.QueryRewriter) (*reg.Profile, error) {
	const q = `
	SELECT id, email, username
	FROM registry.profiles
	WHERE id = @id`

	return r.c.QueryRowContext(ctx, func(r pgx.Row, u *reg.Profile) error {
		return r.Scan(&u.ID, &u.Email, &u.Username)
	}, q, args)
}

type SetProfileArgs struct {
	Email    string
	Username string
	Bio      string
}

func (a SetProfileArgs) RewriteQuery(ctx context.Context, conn *pgx.Conn, sql string, args []any) (newSQL string, newArgs []any, err error) {
	na := pgx.NamedArgs{
		"id":       uuid.New(),
		"email":    a.Email,
		"username": a.Username,
		"bio":      a.Bio,
	}

	return na.RewriteQuery(ctx, conn, sql, args)
}

func (r *repo) SetProfile(ctx context.Context, args pgx.QueryRewriter) error {
	const q = `
	INSERT INTO registry.profiles (id, email, username, bio)
	VALUES (@id, @email, @username, NULLIF(@bio,''))`

	_, err := r.c.ExecContext(ctx, q, args)
	return err
}

func (r *repo) UnsetProfile(ctx context.Context, args pgx.QueryRewriter) error {
	const q = `
	DELETE FROM registry.profiles
	WHERE id = @id`

	count, err := r.c.ExecContext(ctx, q, args)
	if count == 0 {
		return pg.ErrNoRowsAffected
	}

	return err
}

func (r *repo) SetBio(ctx context.Context, args pgx.QueryRewriter) error {
	const q = `
	UPDATE registry.profiles
	SET bio = NULLIF(@bio,'')
	WHERE id = @id`

	count, err := r.c.ExecContext(ctx, q, args)
	if count == 0 {
		return pg.ErrNoRowsAffected
	}

	return err
}

type SetPhotoURLArgs struct {
	ID       uuid.UUID
	PhotoURL string
}

func (a SetPhotoURLArgs) RewriteQuery(ctx context.Context, conn *pgx.Conn, sql string, args []any) (newSQL string, newArgs []any, err error) {
	na := pgx.NamedArgs{
		"id":        a.ID,
		"photo_url": a.PhotoURL,
	}

	return na.RewriteQuery(ctx, conn, sql, args)
}

func (r *repo) SetPhotoURL(ctx context.Context, args pgx.QueryRewriter) error {
	const q = `
	UPDATE registry.profiles
	SET photo_url = NULLIF(@photo_url,'')
	WHERE id = @id`

	count, err := r.c.ExecContext(ctx, q, args)
	if count == 0 {
		return pg.ErrNoRowsAffected
	}

	return err
}

type UpdateProfileArgs struct {
	ID       uuid.UUID
	Username string
	Bio      string
}

func (a UpdateProfileArgs) RewriteQuery(ctx context.Context, conn *pgx.Conn, sql string, args []any) (newSQL string, newArgs []any, err error) {
	na := pgx.NamedArgs{
		"id":       a.ID,
		"username": a.Username,
		"bio":      a.Bio,
	}

	return na.RewriteQuery(ctx, conn, sql, args)
}

func (r *repo) UpdateProfile(ctx context.Context, args pgx.QueryRewriter) error {
	const q = `
	UPDATE registry.profiles
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
	r := &repo{c: pg.NewConn[reg.Profile](rwc)}
	return r
}
