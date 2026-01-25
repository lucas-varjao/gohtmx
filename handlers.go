package main

import (
	"errors"
	"net/http"

	"github.com/angelofallars/htmx-go"

	"github.com/lucas-varjao/gohtmx/internal/middleware"
	"github.com/lucas-varjao/gohtmx/templates"
	"github.com/lucas-varjao/gohtmx/templates/pages"

	"github.com/gin-gonic/gin"
)

// indexViewHandler handles a view for the index page.
func indexViewHandler(c *gin.Context) {

	// Define template meta tags.
	metaTags := pages.MetaTags(
		"gowebly, htmx example page, go with htmx",               // define meta keywords
		"Welcome to example! You're here because it worked out.", // define meta description
	)

	// Define template body content.
	bodyContent := pages.BodyContent(
		"Welcome to example!",                // define h1 text
		"You're here because it worked out.", // define p text
	)

	// Define template layout for index page.
	indexTemplate := templates.Layout(
		"Welcome to example!", // define title text
		metaTags,              // define meta tags
		bodyContent,           // define body content
	)

	// Render index page template.
	if err := htmx.NewResponse().RenderTempl(c.Request.Context(), c.Writer, indexTemplate); err != nil {
		// If not, return HTTP 500 error.
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

}

// showContentAPIHandler handles an API endpoint to show content.
func showContentAPIHandler(c *gin.Context) {
	// Check, if the current request has a 'HX-Request' header.
	// For more information, see https://htmx.org/docs/#request-headers
	if !htmx.IsHTMX(c.Request) {
		// If not, return HTTP 400 error.
		c.AbortWithError(http.StatusBadRequest, errors.New("non-htmx request"))
		return
	}

	// Write HTML content.
	c.Writer.Write([]byte("<p>🎉 Yes, <strong>htmx</strong> is ready to use! (<code>GET /api/hello-world</code>)</p>"))

	// Send htmx response.
	htmx.NewResponse().Write(c.Writer)
}

// loginViewHandler handles a view for the login page.
func loginViewHandler(c *gin.Context) {
	// Check if user is already authenticated, redirect to home
	if sessionID := middleware.ExtractSessionID(c); sessionID != "" {
		c.Redirect(http.StatusFound, "/")
		return
	}

	// Get error message from query parameter if any
	errorMsg := c.Query("error")
	if errorMsg == "" {
		errorMsg = c.GetString("error")
	}

	// Define template meta tags.
	metaTags := pages.MetaTags(
		"login, autenticação, entrar",
		"Faça login na sua conta",
	)

	// Define template body content.
	bodyContent := pages.LoginPage(errorMsg)

	// Define template layout for login page.
	loginTemplate := pages.AuthLayout(
		"Entrar - GoHTMX",
		metaTags,
		bodyContent,
	)

	// Render login page template.
	if err := htmx.NewResponse().RenderTempl(c.Request.Context(), c.Writer, loginTemplate); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
}

// registerViewHandler handles a view for the registration page.
func registerViewHandler(c *gin.Context) {
	// Check if user is already authenticated, redirect to home
	if sessionID := middleware.ExtractSessionID(c); sessionID != "" {
		c.Redirect(http.StatusFound, "/")
		return
	}

	// Get error message from query parameter if any
	errorMsg := c.Query("error")
	if errorMsg == "" {
		errorMsg = c.GetString("error")
	}

	// Define template meta tags.
	metaTags := pages.MetaTags(
		"registro, criar conta, cadastro",
		"Crie uma nova conta",
	)

	// Define template body content.
	bodyContent := pages.RegisterPage(errorMsg)

	// Define template layout for register page.
	registerTemplate := pages.AuthLayout(
		"Criar Conta - GoHTMX",
		metaTags,
		bodyContent,
	)

	// Render register page template.
	if err := htmx.NewResponse().RenderTempl(c.Request.Context(), c.Writer, registerTemplate); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
}
