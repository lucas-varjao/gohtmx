// Package admin provides types and templates for the admin dashboard.
package admin

import "strconv"

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

// DashboardStats holds aggregated user statistics for the admin dashboard.
type DashboardStats struct {
	TotalUsers    int
	ActiveUsers   int
	InactiveUsers int
	AdminUsers    int
	RegularUsers  int
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

// intToString converts an int to string for use in templates.
func intToString(n int) string {
	return strconv.Itoa(n)
}
