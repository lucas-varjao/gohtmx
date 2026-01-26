package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/lucas-varjao/gohtmx/internal/auth"
	"github.com/lucas-varjao/gohtmx/internal/config"
	"github.com/lucas-varjao/gohtmx/internal/handlers"
	"github.com/lucas-varjao/gohtmx/internal/logger"
	"github.com/lucas-varjao/gohtmx/internal/middleware"
	"github.com/lucas-varjao/gohtmx/internal/router"

	"gorm.io/gorm"
)

// TemplRender implements the render.Render interface.
type TemplRender struct {
	Code int
	Data templ.Component
}

// Render implements the render.Render interface.
func (t TemplRender) Render(w http.ResponseWriter) error {
	t.WriteContentType(w)
	w.WriteHeader(t.Code)
	if t.Data != nil {
		return t.Data.Render(context.Background(), w)
	}
	return nil
}

// WriteContentType implements the render.Render interface.
func (t TemplRender) WriteContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
}

// Instance implements the render.Render interface.
func (t *TemplRender) Instance(name string, data any) render.Render {
	_ = name // required by render.Render interface
	if templData, ok := data.(templ.Component); ok {
		return &TemplRender{
			Code: http.StatusOK,
			Data: templData,
		}
	}
	return nil
}

// runServer runs a new HTTP server with the loaded environment variables.
func runServer(authHandler *handlers.AuthHandler, authManager *auth.AuthManager, db *gorm.DB) error {
	cfg := config.GetConfig()
	if cfg == nil {
		return fmt.Errorf("config not loaded")
	}

	// Custom recovery: render HTML error page or JSON depending on Accept header
	recoveryFn := func(c *gin.Context, err any) {
		logger.Error("panic recovered", "error", err)
		if wantsHTML(c) {
			renderErrorPage(c, http.StatusInternalServerError)
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
	}

	// Setup router with all routes (auth, API, etc.)
	r := router.SetupRouter(authHandler, authManager, recoveryFn)

	// Define HTML renderer for template engine (TEMPL support)
	r.HTMLRender = &TemplRender{}

	// Handle static files (keep gowebly static route)
	r.Static("/static", "./static")

	// Handle index page view (receives authManager to show user when logged in)
	r.GET("/", func(c *gin.Context) { indexViewHandler(c, authManager) })

	// Logout from page (invalidates session, clears cookie, redirects to /)
	r.POST("/logout", func(c *gin.Context) { logoutViewHandler(c, authManager) })

	// Handle authentication views (pass authManager for navbar/footer).
	r.GET("/login", func(c *gin.Context) { loginViewHandler(c, authManager) })
	r.GET("/register", func(c *gin.Context) { registerViewHandler(c, authManager) })

	// Handle API endpoints (keep gowebly example route)
	r.GET("/api/hello-world", showContentAPIHandler)

	// Admin area (HTML); requires valid session + admin role
	adminGroup := r.Group("/admin")
	adminGroup.Use(middleware.AdminWebMiddleware(authManager, func(c *gin.Context) { renderErrorPage(c, http.StatusForbidden) }))
	adminGroup.GET("", adminDashboardView)
	adminGroup.GET("/", adminDashboardView)
	adminGroup.GET("/users", func(c *gin.Context) { adminUsersView(c, db, authManager) })
	adminGroup.GET("/users/new", func(c *gin.Context) { adminUsersNewView(c, authManager) })
	adminGroup.POST("/users", func(c *gin.Context) { adminUsersCreatePost(c, db) })
	adminGroup.POST("/users/:id/role", func(c *gin.Context) { adminUserRolePost(c, db) })
	adminGroup.POST("/users/:id/active", func(c *gin.Context) { adminUserActivePost(c, db) })
	adminGroup.POST("/users/:id/delete", func(c *gin.Context) { adminUserDeletePost(c, db, authManager) })

	// 503 maintenance page (for testing and future maintenance mode)
	r.GET("/maintenance", func(c *gin.Context) {
		if wantsHTML(c) {
			renderErrorPage(c, http.StatusServiceUnavailable)
		} else {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		}
	})

	// 404 for unmatched routes (after all other routes)
	r.NoRoute(func(c *gin.Context) {
		if wantsHTML(c) {
			renderErrorPage(c, http.StatusNotFound)
		} else {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "not found"})
		}
	})

	// Get port from config
	port := cfg.Server.Port
	if port == 0 {
		port = 7000 // Default gowebly port
	}

	// Create a new server instance with options from environment variables.
	// For more information, see https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
	// Note: The ReadTimeout and WriteTimeout settings may interfere with SSE (Server-Sent Event) or WS (WebSocket) connections.
	// For SSE or WS, these timeouts can cause the connection to reset after 10 or 5 seconds due to the ReadTimeout and WriteTimeout settings.
	// If you plan to use SSE or WS, consider commenting out or removing the ReadTimeout and WriteTimeout key-value pairs.
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      r,
	}

	// Send log message.
	logger.Info("Starting server...", "port", port)

	return server.ListenAndServe()
}
