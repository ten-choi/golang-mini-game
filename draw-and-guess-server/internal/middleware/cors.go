package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS returns CORS middleware with configured options
func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins: []string{
			"https://studio.apollographql.com",
			"https://sandbox.apollo.dev",
			"http://10.33.255.58:8080",
			"http://localhost:8080",
			"http://localhost:3000",
			"http://localhost:5174",
		},
		AllowMethods: []string{"GET", "POST", "OPTIONS", "PATCH", "PUT", "DELETE"},
		AllowHeaders: []string{
			"Origin", "Content-Type", "Accept", "Authorization",
			"apollo-require-preflight", "x-apollo-operation-name",
			"apollo-client-name", "apollo-client-version",
		},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
