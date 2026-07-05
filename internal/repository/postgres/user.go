package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/irvanmhndra/jejak-api/internal/model"
	"github.com/jmoiron/sqlx"
)

type userRepo struct{ db *sqlx.DB }

func NewUserRepository(db *sqlx.DB) *userRepo { return &userRepo{db} }

func (r *userRepo) Create(ctx context.Context, u *model.User) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, name, email, password_hash) VALUES ($1, $2, $3, $4)`,
		u.ID, u.Name, u.Email, u.PasswordHash,
	)
	return err
}

func (r *userRepo) FindByID(ctx context.Context, id string) (*model.User, error) {
	var u model.User
	err := r.db.GetContext(ctx, &u, `SELECT * FROM users WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &u, err
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	err := r.db.GetContext(ctx, &u, `SELECT * FROM users WHERE email = $1`, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &u, err
}
