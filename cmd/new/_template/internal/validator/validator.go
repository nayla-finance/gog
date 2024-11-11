package validator

import (
	"errors"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

var (
	validate *validator.Validate
	trans    ut.Translator
)

func init() {
	// Initialize validator
	validate = validator.New(validator.WithRequiredStructEnabled())

	// Initialize translator
	english := en.New()
	uni := ut.New(english, english)
	trans, _ = uni.GetTranslator("en")

	// Register default translations
	_ = en_translations.RegisterDefaultTranslations(validate, trans)
}

// Validate performs validation and returns validation errors as a string slice
func Validate(v interface{}) error {
	err := validate.Struct(v)
	if err == nil {
		return nil
	}

	// Convert validation errors to strings
	if validationErrors, ok := err.(validator.ValidationErrors); ok && len(validationErrors) > 0 {
		// we only care about the first error
		return errors.New(validationErrors[0].Translate(trans))
	}

	return err
}
