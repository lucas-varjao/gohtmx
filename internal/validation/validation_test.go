// backend/internal/validation/validation_test.go

package validation

import (
	"strings"
	"testing"
)

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  error
	}{
		{"Valid username", "john_doe123", nil},
		{"Too short", "jo", ErrUsernameTooShort},
		{"Too long", "thisusernameiswaytooooooooooooooooooooooooooooooooooooooooolong", ErrUsernameTooLong},
		{"Empty", "", ErrUsernameInvalid},
		{"Invalid characters", "user@name", ErrUsernameFormat},
		{"Valid with hyphen", "john-doe", nil},
		{"Valid with dot", "john.doe", nil},
		{"Valid with underscore", "john_doe", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsername(tt.username)
			if err != tt.wantErr {
				t.Errorf("ValidateUsername() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr error
	}{
		{"Valid email", "test@example.com", nil},
		{"Empty email", "", ErrEmailInvalid},
		{"No @ symbol", "testexample.com", ErrEmailInvalid},
		{"No domain", "test@", ErrEmailInvalid},
		{"No TLD", "test@example", ErrEmailInvalid},
		{"Valid with plus", "test+tag@example.com", nil},
		{"Valid with dot", "test.name@example.com", nil},
		{"Valid with subdomain", "test@sub.example.com", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if err != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		username string
		wantErr  error
	}{
		{"Valid password", "Test1234!", "", nil},
		{"Too short", "Test12!", "", ErrPasswordTooShort},
		{"No uppercase", "test1234!", "", ErrPasswordNoUppercase},
		{"No lowercase", "TEST1234!", "", ErrPasswordNoLowercase},
		{"No number", "Testabcd!", "", ErrPasswordNoNumber},
		{"No special char", "Test1234", "", ErrPasswordNoSpecial},
		{"Common password", "Password123!", "", ErrPasswordCommonWord},
		{"Contains username", "TestUser123!", "user", ErrPasswordContainsUser},
		{"Complex valid", "C0mpl3x!P@ssw0rd", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password, tt.username)
			if err != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateLoginRequest(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		wantErr  bool
	}{
		{"Valid credentials", "validuser", "password", false},
		{"Empty username", "", "password", true},
		{"Empty password", "validuser", "", true},
		{"Both empty", "", "", true},
		{"Short username", "us", "password", true},
		{"Invalid username chars", "user@name", "password", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLoginRequest(tt.username, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLoginRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRegistrationRequest(t *testing.T) {
	tests := []struct {
		name        string
		username    string
		email       string
		password    string
		displayName string
		wantErr     bool
	}{
		{
			"Valid registration",
			"validuser",
			"valid@example.com",
			"Valid123!",
			"Valid User",
			false,
		},
		{
			"Invalid username",
			"u",
			"valid@example.com",
			"Valid123!",
			"Valid User",
			true,
		},
		{
			"Invalid email",
			"validuser",
			"invalid-email",
			"Valid123!",
			"Valid User",
			true,
		},
		{
			"Weak password",
			"validuser",
			"valid@example.com",
			"weak",
			"Valid User",
			true,
		},
		{
			"Empty display name",
			"validuser",
			"valid@example.com",
			"Valid123!",
			"",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRegistrationRequest(tt.username, tt.email, tt.password, tt.displayName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRegistrationRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateResetToken(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr error
	}{
		{"Valid token", "a1234567890", nil},
		{"Empty token", "", ErrResetTokenInvalid},
		{"Too short", "short", ErrResetTokenInvalid},
		{"Exactly 10 chars", "1234567890", nil},
		{"9 chars is invalid", "123456789", ErrResetTokenInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateResetToken(tt.token)
			if err != tt.wantErr {
				t.Errorf("ValidateResetToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePasswordReset(t *testing.T) {
	tests := []struct {
		name            string
		token           string
		newPassword     string
		confirmPassword string
		wantErr         bool
		errContains     string // optional: substring to check in error message
	}{
		{
			name:            "Valid reset",
			token:           "validtoken12",
			newPassword:     "NewSecure123!",
			confirmPassword: "NewSecure123!",
			wantErr:         false,
		},
		{
			name:            "Invalid token empty",
			token:           "",
			newPassword:     "NewSecure123!",
			confirmPassword: "NewSecure123!",
			wantErr:         true,
			errContains:     "token",
		},
		{
			name:            "Invalid token too short",
			token:           "short",
			newPassword:     "NewSecure123!",
			confirmPassword: "NewSecure123!",
			wantErr:         true,
			errContains:     "token",
		},
		{
			name:            "Passwords do not match",
			token:           "validtoken12",
			newPassword:     "NewSecure123!",
			confirmPassword: "OtherPass123!",
			wantErr:         true,
			errContains:     "coincidem",
		},
		{
			name:            "Weak new password",
			token:           "validtoken12",
			newPassword:     "weak",
			confirmPassword: "weak",
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordReset(tt.token, tt.newPassword, tt.confirmPassword)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePasswordReset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ValidatePasswordReset() error = %v, want message containing %q", err.Error(), tt.errContains)
				}
			}
		})
	}
}
