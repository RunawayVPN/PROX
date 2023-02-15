package main

import (
	"os"
	"time"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/acheong08/ChatGPT-V2/internal/api"
	"github.com/acheong08/ChatGPT-V2/internal/handlers"
	"github.com/fvbock/endless"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var limit_middleware gin.HandlerFunc
var limit_store ratelimit.Store

func init() {
	limit_store = ratelimit.InMemoryStore(
		&ratelimit.InMemoryOptions{
			Rate:  time.Minute,
			Limit: 40,
		},
	)
	limit_middleware = ratelimit.RateLimiter(
		limit_store,
		&ratelimit.Options{
			ErrorHandler: func(c *gin.Context, info ratelimit.Info) {
				c.JSON(
					429,
					gin.H{
						"message": "Too many requests",
					},
				)
				c.Abort()
			},
			KeyFunc: func(c *gin.Context) string {
				// Get Authorization header
				return c.ClientIP()
			},
		},
	)
}

func secret_auth(c *gin.Context) {
	if os.Getenv("SECRET") == "" {
		return
	}
	auth_header := c.GetHeader("Secret")
	if auth_header != os.Getenv("SECRET") {
		c.JSON(401, gin.H{"message": "Unauthorized"})
		c.Abort()
		return
	}
}

func main() {
	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "8080"
	}
	handler := gin.Default()
	if !api.Config.Private {
		handler.Use(limit_middleware)
	}
	// CORS middleware
	cors_config := cors.DefaultConfig()
	cors_config.AllowAllOrigins = true
	cors_config.AllowCredentials = true

	// Set Access-Control-Allow-Credentials headers and allow all origins
	handler.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "duti.tech")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		// Allow ALL CORS
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
		c.Next()
	})

	handler.Use(cors.New(cors_config))

	handler.Use(cors.New(cors_config))
	// handler.Use(secret_auth)
	// Proxy all requests to /* to proxy if not already handled
	handler.Any("/*path", handlers.Proxy)

	endless.ListenAndServe(":"+PORT, handler)
}
