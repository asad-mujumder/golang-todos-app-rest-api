package model

import (
	"time"

	"github.com/google/uuid"
)

type Todo struct {
	ID uuid.UUID `json:"id"`
	Title string `json:"title"`
	Completed bool `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateTodoRequest struct {
	Title string `json:"title" binding:"required,min=1,max=255"`
	Completed bool `json:"completed"`
}