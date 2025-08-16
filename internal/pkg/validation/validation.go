package validation

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}

	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// Validator provides validation methods
type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

// ValidateString validates a string field
func (v *Validator) ValidateString(value, fieldName string, minLength, maxLength int, required bool) error {
	if required && strings.TrimSpace(value) == "" {
		return ValidationError{Field: fieldName, Message: "Обязательное поле"}
	}

	if value != "" {
		length := utf8.RuneCountInString(value)
		if minLength > 0 && length < minLength {
			return ValidationError{Field: fieldName, Message: fmt.Sprintf("Должно быть не менее %d символов", minLength)}
		}
		if maxLength > 0 && length > maxLength {
			return ValidationError{Field: fieldName, Message: fmt.Sprintf("Должно быть не более %d символов", maxLength)}
		}
	}

	return nil
}

// ValidateURL validates a URL field
func (v *Validator) ValidateURL(value, fieldName string, required bool) error {
	if required && strings.TrimSpace(value) == "" {
		return ValidationError{Field: fieldName, Message: "Обязательное поле"}
	}

	if value != "" {
		parsedURL, err := url.Parse(value)
		if err != nil {
			return ValidationError{Field: fieldName, Message: "Неверный URL"}
		}
		if parsedURL.Scheme == "" || parsedURL.Host == "" {
			return ValidationError{Field: fieldName, Message: "URL должен включать схему и хост"}
		}
		if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
			return ValidationError{Field: fieldName, Message: "URL должен использовать схему http или https"}
		}
	}

	return nil
}

// ValidateEmail validates an email field
func (v *Validator) ValidateEmail(value, fieldName string, required bool) error {
	if required && strings.TrimSpace(value) == "" {
		return ValidationError{Field: fieldName, Message: "Обязательное поле"}
	}

	if value != "" {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(value) {
			return ValidationError{Field: fieldName, Message: "Неверный email"}
		}
	}

	return nil
}

// ValidateSlug validates a slug field
func (v *Validator) ValidateSlug(value, fieldName string, required bool) error {
	if required && strings.TrimSpace(value) == "" {
		return ValidationError{Field: fieldName, Message: "Обязательное поле"}
	}

	if value != "" {
		slugRegex := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
		if !slugRegex.MatchString(value) {
			return ValidationError{Field: fieldName, Message: "Slug должен содержать только строчные буквы, цифры и дефисы"}
		}
		if len(value) < 3 {
			return ValidationError{Field: fieldName, Message: "Slug должен быть не менее 3 символов"}
		}
		if len(value) > 50 {
			return ValidationError{Field: fieldName, Message: "Slug должен быть не более 50 символов"}
		}
	}

	return nil
}

// ValidateChoice validates a choice field
func (v *Validator) ValidateChoice(value, fieldName string, choices []string, required bool) error {
	if required && strings.TrimSpace(value) == "" {
		return ValidationError{Field: fieldName, Message: "Обязательное поле"}
	}

	if value != "" {
		found := false
		for _, choice := range choices {
			if value == choice {
				found = true
				break
			}
		}
		if !found {
			return ValidationError{Field: fieldName, Message: fmt.Sprintf("Должно быть одним из: %s", strings.Join(choices, ", "))}
		}
	}

	return nil
}

// ValidateMultiple validates multiple validation rules
func (v *Validator) ValidateMultiple(validations ...error) error {
	var errors ValidationErrors
	for _, validation := range validations {
		if validation != nil {
			if ve, ok := validation.(ValidationError); ok {
				errors = append(errors, ve)
			} else {
				// If it's not a ValidationError, create a generic one
				errors = append(errors, ValidationError{Field: "unknown", Message: validation.Error()})
			}
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}
