// Package handlers provides HTTP request handlers for the API.
package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"

	"github.com/a-h/templ"
	"github.com/lucas-varjao/gohtmx/internal/auth"
	"github.com/lucas-varjao/gohtmx/internal/icons"
	"github.com/lucas-varjao/gohtmx/internal/logger"
	"github.com/lucas-varjao/gohtmx/internal/middleware"
	"github.com/lucas-varjao/gohtmx/internal/service"
	"github.com/lucas-varjao/gohtmx/internal/validation"
	"github.com/lucas-varjao/gohtmx/templates/components"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService service.AuthServiceInterface
}

// renderTemplError renders a templ component as HTML for HTMX error responses.
func renderTemplError(c *gin.Context, component templ.Component) {
	var buf bytes.Buffer
	if err := component.Render(context.Background(), &buf); err != nil {
		logger.Error("Erro ao renderizar componente de erro", "error", err)
		c.String(http.StatusInternalServerError, "Erro ao processar resposta")

		return
	}
	// Determine target based on request path
	target := "#login-error"
	if c.Request.URL.Path == "/auth/register" {
		target = "#register-error"
	}
	c.Header("HX-Retarget", target)
	c.Header("HX-Reswap", "innerHTML")
	// HTMX ignora body em 4xx/5xx; sempre retorne 200 para swap do fragmento.
	c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
}

// renderHTMXError wraps a message with the standard error component.
func renderHTMXError(c *gin.Context, message string) {
	errorAlert := components.ErrorAlert(message, icons.Error())
	renderTemplError(c, errorAlert)
}

// handleLoginBindError logs and responds for binding errors (JSON or HTMX).
func handleLoginBindError(c *gin.Context, err error) {
	logger.Debug("Requisição de login com dados inválidos", "error", err, "ip", getClientIP(c))
	if c.GetHeader("HX-Request") != "" {
		renderHTMXError(c, err.Error())
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

// handleLoginValidationError logs and responds for validation errors (JSON or HTMX).
func handleLoginValidationError(c *gin.Context, req LoginRequest, err error) {
	logger.Debug("Requisição de login com validação falhada", "error", err, "username", req.Username, "ip", getClientIP(c))
	if c.GetHeader("HX-Request") != "" {
		renderHTMXError(c, err.Error())
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

// handleLoginAuthError maps service errors into user-facing responses.
func handleLoginAuthError(c *gin.Context, err error) {
	status := http.StatusUnauthorized
	message := "credenciais inválidas"
	if errors.Is(err, service.ErrUserNotActive) {
		message = "usuário inativo"
	} else if err.Error() == "conta temporariamente bloqueada, tente novamente mais tarde" {
		message = err.Error()
	}

	// HTMX: return 200 so the error fragment is swapped into #login-error (HTMX ignores body on 4xx/5xx)
	if c.GetHeader("HX-Request") != "" {
		renderHTMXError(c, message)
		return
	}

	c.JSON(status, gin.H{"error": message})
}

// getUserAgent safely gets the user agent string from the request.
func getUserAgent(c *gin.Context) string {
	if c.Request == nil {
		return ""
	}
	return c.Request.UserAgent()
}

// setSessionCookie sets the session cookie with consistent flags.
func setSessionCookie(c *gin.Context, sessionID string) {
	// 30 days in seconds.
	const cookieMaxAgeSec = 30 * 24 * 60 * 60
	c.SetCookie(
		middleware.SessionCookieName,
		sessionID,
		cookieMaxAgeSec,
		"/",
		"",
		true, // secure
		true, // httpOnly
	)
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(authService service.AuthServiceInterface) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// LoginRequest represents the login request body (supports both JSON and form data)
type LoginRequest struct {
	Username string `json:"username" binding:"required" form:"username"`
	Password string `json:"password" binding:"required" form:"password"`
}

// RegistrationRequest represents the registration request body (supports both JSON and form data)
type RegistrationRequest struct {
	Username    string `json:"username"     binding:"required" form:"username"`
	Email       string `json:"email"        binding:"required" form:"email"`
	Password    string `json:"password"     binding:"required" form:"password"`
	DisplayName string `json:"display_name" binding:"required" form:"display_name"`
}

// PasswordResetRequest represents the password reset request body
type PasswordResetRequest struct {
	Token           string `json:"token"            binding:"required"`
	NewPassword     string `json:"new_password"     binding:"required"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

// Login handles user authentication with input validation
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	// Support both JSON and form data (for HTMX forms)
	if err := c.ShouldBind(&req); err != nil {
		handleLoginBindError(c, err)
		return
	}

	// Validate input data before attempting login
	if err := validation.ValidateLoginRequest(req.Username, req.Password); err != nil {
		handleLoginValidationError(c, req, err)
		return
	}

	// Get client IP and user agent
	ip := getClientIP(c)
	userAgent := getUserAgent(c)

	response, err := h.authService.Login(req.Username, req.Password, ip, userAgent)
	if err != nil {
		handleLoginAuthError(c, err)
		return
	}

	// Set session cookie for browser sessions.
	setSessionCookie(c, response.SessionID)

	// Check if HTMX request - redirect by role (admin → dashboard, others → home)
	if c.GetHeader("HX-Request") != "" {
		redirectTo := "/"
		if response.User.Role == "admin" {
			redirectTo = "/admin"
		}
		c.Header("HX-Redirect", redirectTo)
		c.Status(http.StatusOK)
		return
	}

	c.JSON(http.StatusOK, response)
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	sessionID, exists := c.Get("sessionID")
	if !exists {
		ip := getClientIP(c)
		logger.Debug("Tentativa de logout sem sessão", "ip", ip)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}

	sessionIDStr := sessionID.(string)
	if err := h.authService.Logout(sessionIDStr); err != nil {
		ip := getClientIP(c)
		logger.Error("Erro ao fazer logout", "error", err, "session_id", sessionIDStr, "ip", ip)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao fazer logout"})
		return
	}

	ip := getClientIP(c)
	logger.Info("Logout realizado com sucesso", "session_id", sessionIDStr, "ip", ip)

	// Clear session cookie
	middleware.ClearSessionCookie(c)

	c.JSON(http.StatusOK, gin.H{"message": "logout realizado com sucesso"})
}

// Register handles new user registration with comprehensive validation
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegistrationRequest
	// Support both JSON and form data (for HTMX forms)
	if err := c.ShouldBind(&req); err != nil {
		logger.Debug("Requisição de registro com dados inválidos", "error", err, "ip", getClientIP(c))
		if c.GetHeader("HX-Request") != "" {
			errorAlert := components.ErrorAlert(err.Error(), icons.Error())
			renderTemplError(c, errorAlert)
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate all registration data
	if err := validation.ValidateRegistrationRequest(
		req.Username,
		req.Email,
		req.Password,
		req.DisplayName,
	); err != nil {
		logger.Debug("Requisição de registro com validação falhada", "error", err, "username", req.Username, "email", req.Email, "ip", getClientIP(c))
		if c.GetHeader("HX-Request") != "" {
			errorAlert := components.ErrorAlert(err.Error(), icons.Error())
			renderTemplError(c, errorAlert)
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Forward to service layer
	user, err := h.authService.Register(req.Username, req.Email, req.Password, req.DisplayName)
	if err != nil {
		logger.Debug("Erro ao registrar usuário", "error", err, "username", req.Username, "email", req.Email, "ip", getClientIP(c))
		if c.GetHeader("HX-Request") != "" {
			errorAlert := components.ErrorAlert(err.Error(), icons.Error())
			renderTemplError(c, errorAlert)
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Strip sensitive data
	user.PasswordHash = ""

	// Check if HTMX request - redirect to login
	if c.GetHeader("HX-Request") != "" {
		c.Header("HX-Redirect", "/login")
		c.Status(http.StatusOK)
		return
	}

	c.JSON(http.StatusOK, user)
}

// RequestPasswordReset handles password reset requests
func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Debug("Requisição de reset de senha com JSON inválido", "error", err, "ip", getClientIP(c))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate email
	if err := validation.ValidateEmail(req.Email); err != nil {
		logger.Debug("Requisição de reset de senha com email inválido", "error", err, "email", req.Email, "ip", getClientIP(c))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authService.RequestPasswordReset(req.Email); err != nil {
		if err.Error() == "invalid email format" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Don't reveal if email exists for security reasons
		c.JSON(http.StatusOK, gin.H{"message": "se o email existir, um link de recuperação será enviado"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "se o email existir, um link de recuperação será enviado"})
}

// ResetPassword handles password reset with token validation
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Debug("Requisição de reset de senha com JSON inválido", "error", err, "ip", getClientIP(c))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate password reset request
	if err := validation.ValidatePasswordReset(req.Token, req.NewPassword, req.ConfirmPassword); err != nil {
		logger.Debug("Requisição de reset de senha com validação falhada", "error", err, "ip", getClientIP(c))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authService.ResetPassword(req.Token, req.NewPassword); err != nil {
		status := http.StatusBadRequest
		ip := getClientIP(c)
		var message string
		switch {
		case errors.Is(err, service.ErrInvalidToken):
			message = "token inválido"
			logger.Warn("Tentativa de reset de senha com token inválido", "ip", ip)
		case errors.Is(err, service.ErrExpiredToken):
			message = "token expirado"
			logger.Warn("Tentativa de reset de senha com token expirado", "ip", ip)
		default:
			message = "falha ao redefinir senha"
			logger.Error("Erro ao resetar senha", "error", err, "ip", ip)
		}
		c.JSON(status, gin.H{"error": message})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "senha redefinida com sucesso"})
}

// GetCurrentUser returns the currently authenticated user
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}

	c.JSON(http.StatusOK, user.(*auth.UserData))
}

// getClientIP safely gets the client IP from the context
// Returns empty string if request is not available (e.g., in tests)
func getClientIP(c *gin.Context) string {
	if c.Request == nil {
		return ""
	}
	return c.ClientIP()
}
