package server

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/ig0rmin/ich/internal/user"
)

func authMiddleware(serverSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			c.Abort()
			return
		}
		tokenString = strings.TrimSpace(tokenString)
		const Bearer = "Bearer "
		if !strings.HasPrefix(tokenString, Bearer) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authroization header format"})
			c.Abort()
		}
		tokenString = strings.TrimPrefix(tokenString, Bearer)

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if token.Method.Alg() != "HS256" {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(serverSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set(user.UserNameKey, claims[user.UserNameKey])
			c.Set(user.UserIDKey, claims[user.UserIDKey])
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad JWT token"})
			c.Abort()
			return
		}

		c.Next()
	}
}
