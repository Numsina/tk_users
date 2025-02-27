package middleware

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Cors() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"x-jwt-token"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://127.0.0.1") {
				return true
			}
			if strings.Contains(origin, "pphcemnbeeohhkhegacmjnocieklpbda") {
				return true
			}
			return strings.Contains(origin, "http://localhost")
		},
		MaxAge: 12 * time.Hour,
	})
}
