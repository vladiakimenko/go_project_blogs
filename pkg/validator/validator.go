package validator

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type HasCustomValidation interface {
	CustomValidate() error
}

type HasPostValidation interface {
	PostValidate() error
}

var ErrTagValidationFailed = errors.New("tag validation failed")
var ErrCustomValidationFailed = errors.New("custom validation failed")
var ErrPostValidationFailed = errors.New("post validation failed")

var validate = validator.New()

// ModelValidate validates struct tags, then calls optional custom validation and post-validation actions in order.
func ModelValidate(value any) error {
	if err := validate.Struct(value); err != nil {
		return fmt.Errorf("%w: %v", ErrTagValidationFailed, err)
	}
	if v, ok := value.(HasCustomValidation); ok {
		if err := v.CustomValidate(); err != nil {
			return fmt.Errorf("%w: %v", ErrCustomValidationFailed, err)
		}
	}
	if v, ok := value.(HasPostValidation); ok {
		if err := v.PostValidate(); err != nil {
			return fmt.Errorf("%w: %v", ErrPostValidationFailed, err)
		}
	}
	return nil
}
