// backend/internal/validation/validation.go

package validation

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var (
	// Error definitions
	ErrUsernameInvalid      = errors.New("nome de usuário inválido")
	ErrUsernameTooShort     = errors.New("nome de usuário deve ter pelo menos 3 caracteres")
	ErrUsernameTooLong      = errors.New("nome de usuário não pode ter mais de 50 caracteres")
	ErrUsernameFormat       = errors.New("nome de usuário pode conter apenas letras, números, pontos, hífens e underscores")
	ErrEmailInvalid         = errors.New("endereço de email inválido")
	ErrPasswordTooShort     = errors.New("senha deve ter pelo menos 8 caracteres")
	ErrPasswordNoUppercase  = errors.New("senha deve conter pelo menos uma letra maiúscula")
	ErrPasswordNoLowercase  = errors.New("senha deve conter pelo menos uma letra minúscula")
	ErrPasswordNoNumber     = errors.New("senha deve conter pelo menos um número")
	ErrPasswordNoSpecial    = errors.New("senha deve conter pelo menos um caractere especial")
	ErrPasswordCommonWord   = errors.New("senha não pode ser uma palavra comum ou fácil de adivinhar")
	ErrPasswordContainsUser = errors.New("senha não pode conter o nome de usuário")
	ErrRefreshTokenInvalid  = errors.New("token de atualização inválido")
	ErrResetTokenInvalid    = errors.New("token de redefinição de senha inválido")
	ErrDisplayNameInvalid   = errors.New("nome de exibição inválido")
	ErrDisplayNameTooLong   = errors.New("nome de exibição não pode ter mais de 100 caracteres")
)

// Validation limits (avoid magic numbers for mnd)
const (
	minUsernameLen = 3
	maxUsernameLen = 50
	minPasswordLen = 8
	maxDisplayLen  = 100
)

// List of common passwords to deny
var commonPasswords = map[string]bool{
	"password":    true,
	"123456":      true,
	"12345678":    true,
	"admin":       true,
	"qwerty":      true,
	"abc123":      true,
	"welcome":     true,
	"welcome1":    true,
	"password123": true,
	"senha123":    true,
}

// ValidateUsername ensures the username meets system requirements
func ValidateUsername(username string) error {
	if username == "" {
		return ErrUsernameInvalid
	}

	if len(username) < minUsernameLen {
		return ErrUsernameTooShort
	}
	if len(username) > maxUsernameLen {
		return ErrUsernameTooLong
	}

	// Username can contain letters, numbers, dots, hyphens, and underscores
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !usernameRegex.MatchString(username) {
		return ErrUsernameFormat
	}

	return nil
}

// ValidateEmail ensures the email format is correct
func ValidateEmail(email string) error {
	if email == "" {
		return ErrEmailInvalid
	}

	// Basic email validation regex
	// For production, consider using a more comprehensive solution or email verification service
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return ErrEmailInvalid
	}

	return nil
}

// ValidatePassword ensures the password meets complexity requirements
func ValidatePassword(password, username string) error {
	if len(password) < minPasswordLen {
		return ErrPasswordTooShort
	}
	if err := validatePasswordChars(password); err != nil {
		return err
	}
	if isCommonPassword(password) {
		return ErrPasswordCommonWord
	}
	if username != "" && len(username) >= minUsernameLen &&
		strings.Contains(strings.ToLower(password), strings.ToLower(username)) {
		return ErrPasswordContainsUser
	}
	return nil
}

func validatePasswordChars(password string) error {
	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, char := range password {
		hasUpper = hasUpper || unicode.IsUpper(char)
		hasLower = hasLower || unicode.IsLower(char)
		hasNumber = hasNumber || unicode.IsNumber(char)
		hasSpecial = hasSpecial || (unicode.IsPunct(char) || unicode.IsSymbol(char))
	}
	if !hasUpper {
		return ErrPasswordNoUppercase
	}
	if !hasLower {
		return ErrPasswordNoLowercase
	}
	if !hasNumber {
		return ErrPasswordNoNumber
	}
	if !hasSpecial {
		return ErrPasswordNoSpecial
	}
	return nil
}

func isCommonPassword(password string) bool {
	lower := strings.ToLower(password)
	for commonPass := range commonPasswords {
		if commonPass == lower || strings.HasPrefix(lower, commonPass) || strings.Contains(lower, commonPass) {
			return true
		}
	}
	return false
}

// ValidateDisplayName validates the display name
func ValidateDisplayName(name string) error {
	if name == "" {
		return ErrDisplayNameInvalid
	}

	if len(name) > maxDisplayLen {
		return ErrDisplayNameTooLong
	}

	return nil
}

// ValidateRefreshToken performs basic validation on refresh tokens
func ValidateRefreshToken(token string) error {
	if token == "" || len(token) < 10 {
		return ErrRefreshTokenInvalid
	}
	return nil
}

// ValidateResetToken performs basic validation on password reset tokens
func ValidateResetToken(token string) error {
	if token == "" || len(token) < 10 {
		return ErrResetTokenInvalid
	}
	return nil
}

// ValidateLoginRequest validates a login request
func ValidateLoginRequest(username, password string) error {
	if err := ValidateUsername(username); err != nil {
		return err
	}

	// For login, we don't apply full password complexity checks
	// since we're only verifying existing credentials
	if password == "" || len(password) < 1 {
		return errors.New("senha não pode ser vazia")
	}

	return nil
}

// ValidateRegistrationRequest validates a registration request
func ValidateRegistrationRequest(username, email, password, displayName string) error {
	if err := ValidateUsername(username); err != nil {
		return err
	}

	if err := ValidateEmail(email); err != nil {
		return err
	}

	if err := ValidatePassword(password, username); err != nil {
		return err
	}

	if err := ValidateDisplayName(displayName); err != nil {
		return fmt.Errorf("nome de exibição inválido: %w", err)
	}

	return nil
}

// ValidatePasswordReset validates a password reset request
func ValidatePasswordReset(token, newPassword, confirmPassword string) error {
	if err := ValidateResetToken(token); err != nil {
		return err
	}

	if newPassword != confirmPassword {
		return errors.New("as senhas não coincidem")
	}

	// For password reset, we don't have username, so use an empty string
	if err := ValidatePassword(newPassword, ""); err != nil {
		return err
	}

	return nil
}
