package token

import (
	"errors"
	"fmt"
)

var (
	ErrTokenInvalid   = errors.New("token is invalid")
	ErrTokenExpired   = errors.New("token has expired")
	ErrTokenMalformed = errors.New("token is malformed")
)

type TokenError struct {
	Op  string
	Err error
}

func (e *TokenError) Error() string {
	return fmt.Sprintf("token %s: %v", e.Op, e.Err)
}

func (e *TokenError) Unwrap() error {
	return e.Err

}
