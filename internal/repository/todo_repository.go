package repository

import (
	"context"
	"fmt"

	"github.com/asad-mujumder/golang-todos-app-rest-api/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type TodoRepository struct {
	pool *pgxpool.Pool
	log zerolog.Logger
}

func NewTodoRepository(pool *pgxpool.Pool, log zerolog.Logger) *TodoRepository {
	return &TodoRepository{
		pool: pool,
		log: log.With().Str("repository", "todo").Logger(),
	}
}

func (r *TodoRepository) Create(reqContext context.Context, title string, completed bool) (*model.Todo, error) {
	ctx, cancel := context.WithTimeout(reqContext, queryTimout)
	defer cancel()
	 
	query := `
	INSERT INTO "todos" ("title", "completed")
	VALUES ($1, $2)
	RETURNING "id", "title", "completed", "created_at", "updated_at"
	`

	r.log.Info().Msg("executing insert query")
	
	var todo model.Todo
	err := r.pool.QueryRow(ctx, query, title, completed).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("todo repository: create: %w", err)
	}

	return &todo, nil
}