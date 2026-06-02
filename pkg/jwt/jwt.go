package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	issuer = "todo_go_api"
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

type Config struct {
	Secret string
	TTL time.Duration
}

type Manager struct {
	cfg Config
}

func NewManager(cfg Config) (*Manager) {
	return &Manager{
		cfg: cfg,
	}
}

func (m *Manager) Generate(userID uuid.UUID) (string, error) {
	claims := Claims{
		userID,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.cfg.TTL)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
			Issuer: issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(m.cfg.Secret))

	if err != nil {
		return "", fmt.Errorf("jwt: generate: sign: %w", err)
	}

	return signed, nil
}

func (m *Manager) Validate(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("jwt: unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(m.cfg.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("jwt: validate: parse: %w", err)
	}

	claims, ok := token.Claims.(*Claims)

	if !ok || !token.Valid {
		return nil, fmt.Errorf("jwt: invalid token")
	}

	return claims, nil
}