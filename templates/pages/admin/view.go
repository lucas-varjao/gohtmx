// Package admin provides types and templates for the admin dashboard.
package admin

// UserView holds display-only user fields for the admin users list.
type UserView struct {
	ID          string
	Username    string
	Email       string
	DisplayName string
	Role        string
	Active      bool
	LastLogin   string
}

// BoolToHidden returns the value to send for the "active" form field when toggling (opposite of current).
func BoolToHidden(active bool) string {
	if active {
		return "false"
	}
	return "true"
}

// BoolToTitle returns the title for the active toggle button.
func BoolToTitle(active bool) string {
	if active {
		return "Clique para desativar"
	}
	return "Clique para ativar"
}
