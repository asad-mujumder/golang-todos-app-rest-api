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

func (r *TodoRepository) List(reqContext context.Context, limit, offset int) ([]*model.Todo, int, error) {
	ctx, cancel := context.WithTimeout(reqContext, queryTimout)
	defer cancel()

	const countQuery = `SELECT COUNT(*) FROM "todos";`

	r.log.Info().Msg("executing count query")
	var total int;
	if err := r.pool.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("todo repository: list: count: %w", err)
	}

	const query = `
	SELECT "id", "title", "completed", "created_at", "updated_at"
	FROM "todos"
	ORDER BY "created_at" DESC 
	LIMIT $1 OFFSET $2;
	`

	r.log.Info().Msg("executing get all query")
	rows, err := r.pool.Query(ctx, query, limit, offset)

	if err != nil {
		return nil, 0, fmt.Errorf("todo repository: list: %w", err)
	}
	defer rows.Close()

	var todos []*model.Todo
	for rows.Next() {
		var todo model.Todo
		if err := rows.Scan(
			&todo.ID,
			&todo.Title,
			&todo.Completed,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("todo repository: list: rows: scan: %w", err)
		}
		todos = append(todos, &todo)
	}

	if err := rows.Err(); err != nil {
			return nil, 0, fmt.Errorf("todo repository: list: rows: %w", err)
	}

	return todos, total, nil
}

func (r *TodoRepository) Create(reqContext context.Context, title string, completed bool) (*model.Todo, error) {
	ctx, cancel := context.WithTimeout(reqContext, queryTimout)
	defer cancel()

	const query = `
	INSERT INTO "todos" ("title", "completed")
	VALUES ($1, $2)
	RETURNING "id", "title", "completed", "created_at", "updated_at";
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