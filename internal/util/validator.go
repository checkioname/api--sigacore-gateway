package util

import (
	"regexp"
	"strings"
)

var (
	_emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	_usernameRegex = regexp.MustCompile(`^[a-z0-9_-]+$`) // Apenas lowercase
)

const (
	ErrUsernameEmpty              = "USERNAME_EMPTY"
	ErrUsernameTooShort           = "USERNAME_TOO_SHORT"
	ErrUsernameTooLong            = "USERNAME_TOO_LONG"
	ErrUsernameInvalidChar        = "USERNAME_INVALID_CHARS"
	ErrUsernameReserved           = "USERNAME_RESERVED"
	ErrUsernameStartsWithNumber   = "USERNAME_STARTS_WITH_NUMBER"
	ErrUsernameConsecutiveSpecial = "USERNAME_CONSECUTIVE_SPECIAL"
)

type UsernameValidationResult struct {
	IsValid      bool
	ErrorMessage string
}

func IsValidUsername(username string) UsernameValidationResult {
	normalized := strings.TrimSpace(strings.ToLower(username))

	if normalized == "" {
		return UsernameValidationResult{false, ErrUsernameEmpty}
	}

	if len(normalized) < 3 {
		return UsernameValidationResult{false, ErrUsernameTooShort}
	}

	if len(normalized) > 50 {
		return UsernameValidationResult{false, ErrUsernameTooLong}
	}

	if !_usernameRegex.MatchString(normalized) {
		return UsernameValidationResult{false, ErrUsernameInvalidChar}
	}

	return UsernameValidationResult{true, ""}
}
