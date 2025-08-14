package auth

import (
	"api--sigacore-gateway/internal/util"

	"github.com/go-playground/validator/v10"
)

var validUsername validator.Func = func(fl validator.FieldLevel) bool {
	if username, ok := fl.Field().Interface().(string); ok {
		return util.IsValidUsername(username).IsValid
	}
	return false
}

var validPassword validator.Func = func(fl validator.FieldLevel) bool {
	if password, ok := fl.Field().Interface().(string); ok {
		if len(password) <= 6 {
			return false
		}

		return true
	}
	return false
}
