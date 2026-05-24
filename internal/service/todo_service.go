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

const (
	defaultLimit = 10
	maxLimit = 50
)

func NewTodoService(repo *repository.TodoRepository, log zerolog.Logger) *TodoService {
	return &TodoService{
		repo: repo,
		log: log.With().Str("service", "todo").Logger(),
	}
}

func (s *TodoService) List(ctx context.Context, req *model.ListTodosRequest) (*model.ListTodosResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}

	if req.Limit < 1 {
		req.Limit = defaultLimit
	}else if req.Limit > maxLimit {
		req.Limit = maxLimit
	}

	offset := (req.Page - 1) * req.Limit

	todos, total, err := s.repo.List(ctx, req.Limit, offset)
	if err != nil {
		return nil, fmt.Errorf("todo service: list: %w", err)
	}

	return &model.ListTodosResponse{
		Todos: todos,
		Total: total,
		Page: req.Page,
		Limit: req.Limit,
	}, nil
}

func (s *TodoService) Create(ctx context.Context, newTodo *model.CreateTodoRequest) (*model.Todo, error) {
	todo, err := s.repo.Create(ctx, newTodo.Title, newTodo.Completed)

	if err != nil {
		return nil, fmt.Errorf("todo service: create: %w", err)
	}

	return  todo, nil
}

func (s *TodoService) Get(ctx context.Context, req *model.GetTodoRequest) (*model.Todo, error) {
	todo, err := s.repo.Get(ctx, req.ID)

	if err != nil {
		return nil, fmt.Errorf("todo service: get: %w", err)
	}

	return todo, nil
}