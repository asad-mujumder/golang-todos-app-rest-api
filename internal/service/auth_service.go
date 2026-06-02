package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/asad-mujumder/golang-todos-app-rest-api/internal/model"
	"github.com/asad-mujumder/golang-todos-app-rest-api/internal/repository"
	"github.com/asad-mujumder/golang-todos-app-rest-api/pkg/jwt"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	repo *repository.AuthRepository
	jwtManager *jwt.Manager
	log zerolog.Logger
}

func NewAuthService(repo *repository.AuthRepository, jwtManager *jwt.Manager, log zerolog.Logger) *AuthService {
	return &AuthService{
		repo: repo,
		jwtManager: jwtManager,
		log: log.With().Str("service", "auth").Logger(),
	}
}

func (s *AuthService) Register(ctx context.Context, req *model.RegisterRequest) (*string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	if err != nil {
		return nil, fmt.Errorf("auth service: register: bcrypt: %w", err)
	}

	newUserID, err := s.repo.Create(ctx, req, string(hashedPassword))

	if err != nil {
		return nil, fmt.Errorf("auth service: register: %w", err)
	}

	return newUserID, nil
}

func (s *AuthService) Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)

	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("auth service: login: %w", err)
	}

	
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, err := s.jwtManager.Generate(user.ID)

	if err != nil {
		return nil, fmt.Errorf("auth service: login: %w", err)
	}

	return &model.LoginResponse{
		User: user,
		AccessToken: accessToken,
	}, nil
}