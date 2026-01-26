// Package router sets up the HTTP routes for the application.
package router

import (
	"net/http"
	"time"

	"github.com/lucas-varjao/gohtmx/internal/auth"
	"github.com/lucas-varjao/gohtmx/internal/handlers"
	"github.com/lucas-varjao/gohtmx/internal/middleware"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// SetupRouter configures all routes for the application.
// If recoveryFn is non-nil, it is used as custom recovery (e.g. to render HTML error pages for 500).
func SetupRouter(
	authHandler *handlers.AuthHandler,
	authManager *auth.AuthManager,
	recoveryFn gin.RecoveryFunc,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	if recoveryFn != nil {
		r.Use(gin.CustomRecovery(recoveryFn))
	} else {
		r.Use(gin.Recovery())
	}

	// Add CORS middleware
	r.Use(middleware.CorsMiddleware())

	// Health check routes
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// Rate limiter for auth routes (brute force prevention)
	const authBurst = 3
	authLimiter := middleware.NewIPRateLimiter(rate.Limit(1), authBurst, time.Hour)

	// Public auth routes
	authRoutes := r.Group("/auth")
	authRoutes.Use(middleware.RateLimitMiddleware(authLimiter))
	authRoutes.POST("/login", authHandler.Login)
	authRoutes.POST("/register", authHandler.Register)
	authRoutes.POST("/password-reset-request", authHandler.RequestPasswordReset)
	authRoutes.POST("/password-reset", authHandler.ResetPassword)

	// Rate limiter for API (more permissive)
	const apiBurst = 20
	const apiRatePerSec = 10
	apiLimiter := middleware.NewIPRateLimiter(rate.Limit(apiRatePerSec), apiBurst, time.Hour)

	// Protected routes
	api := r.Group("/api")
	api.Use(middleware.RateLimitMiddleware(apiLimiter))
	api.Use(middleware.AuthMiddleware(authManager))
	api.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Esta é uma rota protegida"})
	})
	api.GET("/me", authHandler.GetCurrentUser)
	api.POST("/logout", authHandler.Logout)

	// Admin only routes
	admin := api.Group("/admin")
	admin.Use(middleware.RoleMiddleware("admin"))
	admin.GET("/dashboard", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Admin Dashboard"})
	})

	return r
}
