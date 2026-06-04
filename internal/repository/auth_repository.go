package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/asad-mujumder/golang-todos-app-rest-api/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type AuthRepository struct {
	pool *pgxpool.Pool;
	log zerolog.Logger
}

func NewAuthRepository(pool *pgxpool.Pool, log zerolog.Logger) *AuthRepository {
	return &AuthRepository{
		pool: pool,
		log: log.With().Str("repository", "auth").Logger(),
	}
}

func (r *AuthRepository) Create(reqContext context.Context, req *model.RegisterRequest, hashedPassword string) (*string, error) {
	ctx, cancel := context.WithTimeout(reqContext, queryTimout)
	defer cancel()

	const query = `
		INSERT INTO "users" ("first_name", "last_name", "email", "password")
		VALUES ($1, $2, $3, $4)
		RETURNING "id";
	`

	r.log.Info().Msg("executing insert query")
	var newUserID string
	err := r.pool.QueryRow(ctx, query, req.FirstName, req.LastName, req.Email, hashedPassword).Scan(&newUserID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicateEmail
		}
		return nil, fmt.Errorf("auth repository: create: %w", err)
	}

	return &newUserID, nil
}

func (r *AuthRepository) GetByEmail(reqContext context.Context, email string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(reqContext, queryTimout)
	defer cancel()

	const query = `
		SELECT "id", "first_name", "last_name", "email", "password", "created_at", "updated_at"
		FROM "users"
		WHERE "email" = $1;
	`

	r.log.Info().Msg("executing get by email query")
	var user model.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("auth repository: get by email: %w", err)
	}

	return &user, nil
}