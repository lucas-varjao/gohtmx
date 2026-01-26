// Package icons provides shared Lucide icon helpers for consistent styling across the app.
// Icons are returned as template.HTML for use in TEMPL via @templ.Raw(icon).
// To add icons: use lucide.Icon("name", opts) or lucide.IconName(lucide.Options{...}).
// Icon names: https://lucide.dev/icons. API: https://github.com/kaugesaar/lucide-go.
package icons

import (
	"html/template"

	"github.com/kaugesaar/lucide-go"
)

// Class names used for different contexts.
const (
	classAlert  = "stroke-current shrink-0 h-6 w-6"
	classButton = "w-4 h-4 shrink-0"
	classLabel  = "w-4 h-4 shrink-0 opacity-70"
)

// currentColor makes the SVG stroke inherit text color. lucide-go's per-icon helpers
// do not set a default, so we set it explicitly to avoid stroke="" (invisible icons).
const colorCurrent = "currentColor"

// Error returns the CircleX icon for error alerts (DaisyUI alert-error).
func Error() template.HTML {
	return lucide.CircleX(lucide.Options{Color: colorCurrent, Class: classAlert})
}

// LogIn returns the log-in icon for the "Entrar" navbar link and login submit button.
func LogIn() template.HTML {
	return lucide.LogIn(lucide.Options{Color: colorCurrent, Class: classButton})
}

// LogOut returns the log-out icon for the "Sair" navbar button.
func LogOut() template.HTML {
	return lucide.LogOut(lucide.Options{Color: colorCurrent, Class: classButton})
}

// UserPlus returns the user-plus icon for "Registrar" and "Criar Conta" submit button.
func UserPlus() template.HTML {
	return lucide.UserPlus(lucide.Options{Color: colorCurrent, Class: classButton})
}

// User returns the user icon for username/identifier form labels.
func User() template.HTML {
	return lucide.User(lucide.Options{Color: colorCurrent, Class: classLabel})
}

// Mail returns the mail icon for email form labels.
func Mail() template.HTML {
	return lucide.Mail(lucide.Options{Color: colorCurrent, Class: classLabel})
}

// Lock returns the lock icon for password form labels.
func Lock() template.HTML {
	return lucide.Lock(lucide.Options{Color: colorCurrent, Class: classLabel})
}

// UserCircle returns the user-round icon for display-name form labels.
func UserCircle() template.HTML {
	return lucide.CircleUser(lucide.Options{Color: colorCurrent, Class: classLabel})
}

// ValidationSuccess returns the CircleCheck icon for “requisito atendido” in password validation.
// Small size (w-3 h-3) to match discreet validation text.
func ValidationSuccess() template.HTML {
	return lucide.CircleCheck(lucide.Options{Color: colorCurrent, Class: "w-2 h-2 shrink-0"})
}

// ValidationFail returns the CircleX icon for “requisito não atendido” in password validation.
// Small size (w-3 h-3) to match discreet validation text.
func ValidationFail() template.HTML {
	return lucide.CircleX(lucide.Options{Color: colorCurrent, Class: "w-2 h-2 shrink-0"})
}
