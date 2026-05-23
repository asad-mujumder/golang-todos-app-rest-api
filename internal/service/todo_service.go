package service

import (
	"context"
	"fmt"

	"github.com/asad-mujumder/golang-todos-app-rest-api/internal/model"
	"github.com/asad-mujumder/golang-todos-app-rest-api/internal/repository"
	"github.com/rs/zerolog"
)

type TodoService struct {
	repo *repository.TodoRepository
	log zerolog.Logger
}

func NewTodoService(repo *repository.TodoRepository, log zerolog.Logger) *TodoService {
	return &TodoService{
		repo: repo,
		log: log.With().Str("service", "todo").Logger(),
	}
}

func (s *TodoService) Create(ctx context.Context, newTodo *model.CreateTodoRequest) (*model.Todo, error) {
	todo, err := s.repo.Create(ctx, newTodo.Title, newTodo.Completed)

	if err != nil {
		return nil, fmt.Errorf("todo service: create: %w", err)
	}

	return  todo, nil
}