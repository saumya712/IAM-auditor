package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware sets CORS headers on every response.
// The allowed origin is read from the FRONTEND_ORIGIN env var (default "*").
// OPTIONS preflight requests are answered with 204 No Content.
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := os.Getenv("FRONTEND_ORIGIN")
		if origin == "" {
			origin = "*"
		}

		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
