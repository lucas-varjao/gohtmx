package main

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/angelofallars/htmx-go"

	"github.com/lucas-varjao/gohtmx/internal/auth"
	"github.com/lucas-varjao/gohtmx/internal/icons"
	"github.com/lucas-varjao/gohtmx/internal/middleware"
	"github.com/lucas-varjao/gohtmx/internal/models"
	"github.com/lucas-varjao/gohtmx/internal/validation"
	"github.com/lucas-varjao/gohtmx/templates/components"
	"github.com/lucas-varjao/gohtmx/templates/layouts"
	"github.com/lucas-varjao/gohtmx/templates/pages"
	"github.com/lucas-varjao/gohtmx/templates/pages/admin"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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
		icons.LogIn(),
		icons.UserPlus(),
		icons.LogOut(),
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
	bodyContent := layouts.AuthContentWrap(pages.LoginPage(errorMsg, icons.Error(), icons.LogIn(), icons.User(), icons.Lock()))

	loginTemplate := layouts.Layout(
		"Entrar - GoHTMX",
		metaTags,
		bodyContent,
		displayName,
		loggedIn,
		icons.LogIn(),
		icons.UserPlus(),
		icons.LogOut(),
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
	bodyContent := layouts.AuthContentWrap(pages.RegisterPage(errorMsg, icons.Error(), icons.UserPlus(), icons.User(), icons.Mail(), icons.UserCircle(), icons.Lock(), icons.ValidationSuccess(), icons.ValidationFail()))

	registerTemplate := layouts.Layout(
		"Criar Conta - GoHTMX",
		metaTags,
		bodyContent,
		displayName,
		loggedIn,
		icons.LogIn(),
		icons.UserPlus(),
		icons.LogOut(),
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

// adminDashboardView redirects to the users list (main admin view).
func adminDashboardView(c *gin.Context) {
	c.Redirect(http.StatusFound, "/admin/users")
}

// adminUsersView renders the admin users list inside the dashboard layout.
func adminUsersView(c *gin.Context, db *gorm.DB) {
	var users []models.User
	if err := db.Order("created_at DESC").Find(&users).Error; err != nil {
		renderErrorPage(c, http.StatusInternalServerError)
		return
	}
	views := make([]admin.UserView, 0, len(users))
	for _, u := range users {
		lastLogin := ""
		if !u.LastLogin.IsZero() {
			lastLogin = u.LastLogin.Format("02/01/2006 15:04")
		}
		views = append(views, admin.UserView{
			ID:          strconv.FormatUint(uint64(u.ID), 10),
			Username:    u.Username,
			Email:       u.Email,
			DisplayName: u.DisplayName,
			Role:        u.Role,
			Active:      u.Active,
			LastLogin:   lastLogin,
		})
	}
	metaTags := pages.MetaTags("admin, usuários, gestão", "Gerencie usuários do sistema.")
	content := admin.UsersPage(views, icons.CircleCheckForStatus(), icons.ValidationFail(), icons.Trash2(), icons.Error())
	tmpl := layouts.DashboardLayout(
		"Usuários - Admin - GoHTMX",
		metaTags,
		"users",
		icons.LayoutDashboard(),
		icons.Users(),
		icons.LogOut(),
		content,
	)
	if err := htmx.NewResponse().RenderTempl(c.Request.Context(), c.Writer, tmpl); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}

// userViewFromModel converts a models.User to admin.UserView (ID as string, last login formatted).
func userViewFromModel(u *models.User) admin.UserView {
	lastLogin := ""
	if !u.LastLogin.IsZero() {
		lastLogin = u.LastLogin.Format("02/01/2006 15:04")
	}
	return admin.UserView{
		ID:          strconv.FormatUint(uint64(u.ID), 10),
		Username:    u.Username,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Role:        u.Role,
		Active:      u.Active,
		LastLogin:   lastLogin,
	}
}

// adminUserRolePost updates a user's role and returns the updated table row HTML for HTMX swap.
func adminUserRolePost(c *gin.Context, db *gorm.DB) {
	idStr := c.Param("id")
	// PostForm reads from both URL query and body; form from HTMX is in body as application/x-www-form-urlencoded
	role := c.PostForm("role")
	if role == "" {
		role = c.Request.PostFormValue("role")
	}
	if role != "admin" && role != "user" {
		role = "user"
	}
	var u models.User
	if err := db.First(&u, idStr).Error; err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	u.Role = role
	if err := db.Model(&u).Update("role", role).Error; err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	view := userViewFromModel(&u)
	row := admin.UserRow(view, icons.CircleCheckForStatus(), icons.ValidationFail(), icons.Trash2())
	c.Header("Content-Type", "text/html; charset=utf-8")
	_ = row.Render(context.Background(), c.Writer)
}

// adminUserActivePost toggles a user's active status and returns the updated table row HTML for HTMX swap.
func adminUserActivePost(c *gin.Context, db *gorm.DB) {
	idStr := c.Param("id")
	activeStr := c.PostForm("active")
	active := activeStr == "true" || activeStr == "1"
	var u models.User
	if err := db.First(&u, idStr).Error; err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	u.Active = active
	if err := db.Save(&u).Error; err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	view := userViewFromModel(&u)
	row := admin.UserRow(view, icons.CircleCheckForStatus(), icons.ValidationFail(), icons.Trash2())
	c.Header("Content-Type", "text/html; charset=utf-8")
	_ = row.Render(context.Background(), c.Writer)
}

// adminUserDeletePost permanently deletes a user (hard delete), clears their sessions, then redirects to /admin/users.
func adminUserDeletePost(c *gin.Context, db *gorm.DB, authManager *auth.AuthManager) {
	idStr := c.Param("id")
	var u models.User
	if err := db.First(&u, idStr).Error; err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	userID := strconv.FormatUint(uint64(u.ID), 10)
	_ = authManager.LogoutAll(userID)
	if err := db.Unscoped().Delete(&u).Error; err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if c.GetHeader("HX-Request") != "" {
		c.Header("HX-Redirect", "/admin/users")
		c.Status(http.StatusOK)
		return
	}
	c.Redirect(http.StatusFound, "/admin/users")
}

// adminUsersNewView renders the new-user form inside the dashboard layout.
func adminUsersNewView(c *gin.Context) {
	errorMsg := c.Query("error")
	if errorMsg == "" {
		errorMsg = c.GetString("error")
	}
	metaTags := pages.MetaTags("admin, novo usuário, criar conta", "Criar novo usuário")
	content := admin.UsersNewPage(errorMsg, icons.Error())
	tmpl := layouts.DashboardLayout(
		"Novo usuário - Admin - GoHTMX",
		metaTags,
		"users",
		icons.LayoutDashboard(),
		icons.Users(),
		icons.LogOut(),
		content,
	)
	if err := htmx.NewResponse().RenderTempl(c.Request.Context(), c.Writer, tmpl); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}

// adminUsersCreatePost creates a user from the form and redirects to /admin/users (or returns error fragment for HTMX).
func adminUsersCreatePost(c *gin.Context, db *gorm.DB) {
	username := c.PostForm("username")
	email := c.PostForm("email")
	displayName := c.PostForm("display_name")
	password := c.PostForm("password")
	role := c.PostForm("role")
	if role != "admin" && role != "user" {
		role = "user"
	}
	active := c.PostForm("active") == "true" || c.PostForm("active") == "1"

	if err := validation.ValidateRegistrationRequest(username, email, password, displayName); err != nil {
		if c.GetHeader("HX-Request") != "" {
			// HTMX não faz swap em 4xx; retornar 200 para o conteúdo de erro ser colocado em #new-user-error
			alert := components.ErrorAlert(err.Error(), icons.Error())
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.Header("HX-Retarget", "#new-user-error")
			c.Header("HX-Reswap", "innerHTML")
			c.Status(http.StatusOK)
			_ = alert.Render(context.Background(), c.Writer)
			return
		}
		c.Redirect(http.StatusSeeOther, "/admin/users/new?error="+url.QueryEscape(err.Error()))
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		renderErrorPage(c, http.StatusInternalServerError)
		return
	}
	u := models.User{
		Username:     username,
		Email:        email,
		DisplayName:  displayName,
		PasswordHash: string(hashedPassword),
		Role:         role,
		Active:       active,
	}
	if err := db.Create(&u).Error; err != nil {
		msg := "usuário ou email já existe"
		if c.GetHeader("HX-Request") != "" {
			// HTMX não faz swap em 4xx; retornar 200 para o conteúdo de erro ser colocado em #new-user-error
			alert := components.ErrorAlert(msg, icons.Error())
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.Header("HX-Retarget", "#new-user-error")
			c.Header("HX-Reswap", "innerHTML")
			c.Status(http.StatusOK)
			_ = alert.Render(context.Background(), c.Writer)
			return
		}
		c.Redirect(http.StatusSeeOther, "/admin/users/new?error="+url.QueryEscape(msg))
		return
	}
	if c.GetHeader("HX-Request") != "" {
		c.Header("HX-Redirect", "/admin/users")
		c.Status(http.StatusOK)
		return
	}
	c.Redirect(http.StatusFound, "/admin/users")
}
