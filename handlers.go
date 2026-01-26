package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/angelofallars/htmx-go"

	"github.com/lucas-varjao/gohtmx/internal/auth"
	"github.com/lucas-varjao/gohtmx/internal/middleware"
	"github.com/lucas-varjao/gohtmx/templates"
	"github.com/lucas-varjao/gohtmx/templates/pages"

	"github.com/gin-gonic/gin"
)

// AppVersion is shown in the footer. Set via ldflags on release or use "dev".
var AppVersion = "dev"

// getNavData returns displayName and loggedIn for the navbar from the current request.
func getNavData(c *gin.Context, authManager *auth.AuthManager) (displayName string, loggedIn bool) {
	sessionID := middleware.ExtractSessionID(c)
	if sessionID == "" {
		return "", false
	}
	_, user, err := authManager.ValidateSession(sessionID)
	if err != nil || user == nil {
		return "", false
	}
	loggedIn = true
	if user.DisplayName != "" {
		displayName = user.DisplayName
	} else {
		displayName = user.Identifier
	}
	return displayName, loggedIn
}

// indexViewHandler handles the index page; shows user name + logout when logged in.
func indexViewHandler(c *gin.Context, authManager *auth.AuthManager) {
	displayName, loggedIn := getNavData(c, authManager)
	generatedAt := time.Now().Format("02/01/2006 15:04:05")

	metaTags := pages.MetaTags(
		"GoHTMX, Go, TEMPL, HTMX, Alpine.js, Tailwind, DaisyUI, demo, stack",
		"Página de demonstração da stack: Go, TEMPL, HTMX, Alpine.js, Tailwind e DaisyUI.",
	)

	bodyContent := pages.IndexPage(generatedAt)

	indexTemplate := templates.Layout(
		"GoHTMX — Stack demo",
		metaTags,
		bodyContent,
		displayName,
		loggedIn,
		AppVersion,
		time.Now().Year(),
	)

	if err := htmx.NewResponse().RenderTempl(c.Request.Context(), c.Writer, indexTemplate); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
}

// logoutViewHandler invalidates the session and redirects to index.
func logoutViewHandler(c *gin.Context, authManager *auth.AuthManager) {
	sessionID := middleware.ExtractSessionID(c)
	if sessionID != "" {
		_ = authManager.Logout(sessionID)
		middleware.ClearSessionCookie(c)
	}
	c.Redirect(http.StatusFound, "/")
}

// showContentAPIHandler handles an API endpoint to show content.
func showContentAPIHandler(c *gin.Context) {
	// Check, if the current request has a 'HX-Request' header.
	// For more information, see https://htmx.org/docs/#request-headers
	if !htmx.IsHTMX(c.Request) {
		// If not, return HTTP 400 error.
		_ = c.AbortWithError(http.StatusBadRequest, errors.New("non-htmx request"))
		return
	}

	// Write HTML content.
	_, _ = c.Writer.WriteString("<p>🎉 Yes, <strong>htmx</strong> is ready to use! (<code>GET /api/hello-world</code>)</p>")

	// Send htmx response.
	_ = htmx.NewResponse().Write(c.Writer)
}

// loginViewHandler handles a view for the login page.
func loginViewHandler(c *gin.Context, authManager *auth.AuthManager) {
	if sessionID := middleware.ExtractSessionID(c); sessionID != "" {
		c.Redirect(http.StatusFound, "/")
		return
	}

	errorMsg := c.Query("error")
	if errorMsg == "" {
		errorMsg = c.GetString("error")
	}

	displayName, loggedIn := getNavData(c, authManager)
	metaTags := pages.MetaTags("login, autenticação, entrar", "Faça login na sua conta")
	bodyContent := pages.AuthContentWrap(pages.LoginPage(errorMsg))

	loginTemplate := templates.Layout(
		"Entrar - GoHTMX",
		metaTags,
		bodyContent,
		displayName,
		loggedIn,
		AppVersion,
		time.Now().Year(),
	)

	if err := htmx.NewResponse().RenderTempl(c.Request.Context(), c.Writer, loginTemplate); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
}

// registerViewHandler handles a view for the registration page.
func registerViewHandler(c *gin.Context, authManager *auth.AuthManager) {
	if sessionID := middleware.ExtractSessionID(c); sessionID != "" {
		c.Redirect(http.StatusFound, "/")
		return
	}

	errorMsg := c.Query("error")
	if errorMsg == "" {
		errorMsg = c.GetString("error")
	}

	displayName, loggedIn := getNavData(c, authManager)
	metaTags := pages.MetaTags("registro, criar conta, cadastro", "Crie uma nova conta")
	bodyContent := pages.AuthContentWrap(pages.RegisterPage(errorMsg))

	registerTemplate := templates.Layout(
		"Criar Conta - GoHTMX",
		metaTags,
		bodyContent,
		displayName,
		loggedIn,
		AppVersion,
		time.Now().Year(),
	)

	if err := htmx.NewResponse().RenderTempl(c.Request.Context(), c.Writer, registerTemplate); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
}
