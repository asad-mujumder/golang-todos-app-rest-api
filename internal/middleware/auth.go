package middleware

import (
	"net/http"
	"strings"

	"github.com/asad-mujumder/golang-todos-app-rest-api/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

const UserIDKey = "user_id"

func Auth(jwtManager *jwt.Manager, log zerolog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "missing authorization header"})
            c.Abort()
            return
        }

        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "invalid authorization header"})
            c.Abort()
            return
        }

        claims, err := jwtManager.Validate(parts[1])
        if err != nil {
            log.Warn().Err(err).Msg("invalid token")
            c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "invalid or expired token"})
            c.Abort()
            return
        }

        if newToken, err := jwtManager.Generate(claims.UserID); err == nil {
            log.Info().Msg(newToken)
            c.Header("X-NEW-ACCESS-TOKEN", newToken)
        }
            
        c.Set(UserIDKey, claims.UserID)
        c.Next()
    }
}