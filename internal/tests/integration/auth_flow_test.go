// Package integration provides integration tests for the authentication flow.
package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lucas-varjao/gohtmx/internal/auth"
	gormadapter "github.com/lucas-varjao/gohtmx/internal/auth/adapter/gorm"
	"github.com/lucas-varjao/gohtmx/internal/email"
	"github.com/lucas-varjao/gohtmx/internal/handlers"
	"github.com/lucas-varjao/gohtmx/internal/models"
	"github.com/lucas-varjao/gohtmx/internal/router"
	"github.com/lucas-varjao/gohtmx/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupIntegrationTest(t *testing.T) (*gin.Engine, *gorm.DB, *auth.AuthManager) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&models.User{}, &models.Session{})
	require.NoError(t, err)

	// Setup adapters
	userAdapter := gormadapter.NewUserAdapter(db)
	sessionAdapter := gormadapter.NewSessionAdapter(db)

	// Setup auth manager
	authConfig := auth.DefaultAuthConfig()
	authManager := auth.NewAuthManager(userAdapter, sessionAdapter, authConfig)

	// Setup services
	emailService := email.NewMockEmailService()
	authService := service.NewAuthService(authManager, userAdapter, emailService)
	authHandler := handlers.NewAuthHandler(authService)

	// Setup router
	r := router.SetupRouter(authHandler, authManager, nil)
	return r, db, authManager
}

func TestCompleteAuthFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, _, _ := setupIntegrationTest(t)

	// 1. Register user
	registration := map[string]any{
		"username":     "testuser",
		"email":        "test@example.com",
		"password":     "Test123!@#",
		"display_name": "Test User",
	}
	w := httptest.NewRecorder()
	jsonData, _ := json.Marshal(registration)
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 2. Login
	login := map[string]any{
		"username": "testuser",
		"password": "Test123!@#",
	}
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(login)
	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var loginResponse map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &loginResponse)
	require.NoError(t, err)
	sessionID := loginResponse["session_id"].(string)
	assert.NotEmpty(t, sessionID)

	// 3. Access protected route
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer "+sessionID)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 4. Logout
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/logout", nil)
	req.Header.Set("Authorization", "Bearer "+sessionID)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 5. Attempt access after logout (should fail)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer "+sessionID)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPasswordResetFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, db, _ := setupIntegrationTest(t)

	// 1. Create user directly in database
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("oldpassword123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	user := &models.User{
		Username:     "resetuser",
		Email:        "reset@example.com",
		PasswordHash: string(hashedPassword),
		DisplayName:  "Reset User",
		Active:       true,
		Role:         "user",
	}
	err = db.Create(user).Error
	require.NoError(t, err)

	// 2. Request password reset
	resetRequest := map[string]any{
		"email": "reset@example.com",
	}
	w := httptest.NewRecorder()
	jsonData, _ := json.Marshal(resetRequest)
	req, _ := http.NewRequest("POST", "/auth/password-reset-request", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify reset token was set in the in-memory DB (option B)
	var updatedUser models.User
	err = db.First(&updatedUser, user.ID).Error
	require.NoError(t, err)
	assert.NotEmpty(t, updatedUser.ResetToken, "reset token should be set after password-reset-request")
	assert.False(t, updatedUser.ResetTokenExpiry.IsZero(), "reset token expiry should be set")
}

func TestGetCurrentUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, _, _ := setupIntegrationTest(t)

	// 1. Register and login
	registration := map[string]any{
		"username":     "meuser",
		"email":        "me@example.com",
		"password":     "Test123!@#",
		"display_name": "Me User",
	}
	w := httptest.NewRecorder()
	jsonData, _ := json.Marshal(registration)
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	login := map[string]any{
		"username": "meuser",
		"password": "Test123!@#",
	}
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(login)
	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	var loginResponse map[string]any
	json.Unmarshal(w.Body.Bytes(), &loginResponse)
	sessionID := loginResponse["session_id"].(string)

	// 2. Get current user
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/me", nil)
	req.Header.Set("Authorization", "Bearer "+sessionID)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var userResponse map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &userResponse)
	require.NoError(t, err)
	assert.Equal(t, "meuser", userResponse["identifier"])
}
