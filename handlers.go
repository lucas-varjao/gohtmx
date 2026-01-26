package main

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/angelofallars/htmx-go"

	"github.com/lucas-varjao/gohtmx/internal/auth"
	"github.com/lucas-varjao/gohtmx/internal/middleware"
	"github.com/lucas-varjao/gohtmx/templates/layouts"
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

	indexTemplate := layouts.Layout(
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
	bodyContent := layouts.AuthContentWrap(pages.LoginPage(errorMsg))

	loginTemplate := layouts.Layout(
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
	bodyContent := layouts.AuthContentWrap(pages.RegisterPage(errorMsg))

	registerTemplate := layouts.Layout(
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

// wantsHTML returns true when the request prefers an HTML response (browser navigation).
func wantsHTML(c *gin.Context) bool {
	accept := c.GetHeader("Accept")
	if accept == "" {
		return true
	}
	return strings.Contains(accept, "text/html")
}

// renderErrorPage writes an HTTP error page (404, 403, 500, 503) using the error layout and template.
func renderErrorPage(c *gin.Context, code int) {
	var title string
	var metaKeywords, metaDesc string
	var content templ.Component
	switch code {
	case http.StatusNotFound:
		title = "Página não encontrada - GoHTMX"
		metaKeywords = "erro 404, não encontrado"
		metaDesc = "Página não encontrada."
		content = pages.Error404Content()
	case http.StatusForbidden:
		title = "Acesso negado - GoHTMX"
		metaKeywords = "erro 403, acesso negado"
		metaDesc = "Acesso negado."
		content = pages.Error403Content()
	case http.StatusInternalServerError:
		title = "Erro interno - GoHTMX"
		metaKeywords = "erro 500, erro interno"
		metaDesc = "Erro interno do servidor."
		content = pages.Error500Content()
	case http.StatusServiceUnavailable:
		title = "Em manutenção - GoHTMX"
		metaKeywords = "erro 503, manutenção"
		metaDesc = "Serviço temporariamente indisponível."
		content = pages.Error503Content()
	default:
		code = http.StatusInternalServerError
		title = "Erro - GoHTMX"
		metaKeywords = "erro"
		metaDesc = "Ocorreu um erro."
		content = pages.Error500Content()
	}
	metaTags := pages.MetaTags(metaKeywords, metaDesc)
	tmpl := layouts.ErrorLayout(title, metaTags, content)
	c.Status(code)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Render(context.Background(), c.Writer); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}
