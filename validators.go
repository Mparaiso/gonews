package gonews

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// ValidationError holds errors in a map
type ValidationError map[string][]string

// Append adds an new error to a map
func (v ValidationError) Append(key string, value string) {
	v[key] = append(v[key], value)
}

func (v ValidationError) Error() string {
	return fmt.Sprintf("%#v", v)
}

// HasErrors returns true if error exists
func (v ValidationError) HasErrors() bool {
	return len(v) != 0
}

// UserValidator is a User validator
type UserValidator struct {
}

// Validate validates a user
func (uv UserValidator) Validate(u *User) ValidationError {
	errors := ValidationError{}
	if strings.Trim(u.Username, " ") == "" {
		errors.Append("Username", "should not be empty")
	}
	if len(u.Username) < 6 {
		errors.Append("Username", "should be at least 6 characters long")
	}
	if len(u.Username) > 100 {
		errors.Append("Username", "should be at most 100 characters long")
	}
	if strings.Trim(u.Email, " ") == "" {
		errors.Append("Email", "should not be empty")
	}
	if len(u.Email) < 6 {
		errors.Append("Email", "should be at least 6 characters long")
	}
	if len(u.Email) > 100 {
		errors.Append("Email", "should be at most 100 characters long")
	}
	if !isEmail(u.Email) {
		errors.Append("Email", "should be a valid email")
	}
	if strings.Trim(u.Password, " ") == "" {
		errors.Append("Password", "should not be empty")
	}

	if !errors.HasErrors() {
		return nil
	}
	return errors
}

func isEmail(candidate string) bool {
	return regexp.MustCompile(`\w+@\w+\.\w+`).MatchString(candidate)
}

// RegistrationFormValidator is a RegistrationForm validator
type RegistrationFormValidator struct {
	request        *http.Request
	csrfProvider   CSRFProvider
	userRepository UserFinder
}

// NewRegistrationFormValidator creates an new RegistrationFormValidator
func NewRegistrationFormValidator(request *http.Request, csrfProvider CSRFProvider, userFinder UserFinder) *RegistrationFormValidator {
	return &RegistrationFormValidator{request, csrfProvider, userFinder}
}

// Validate returns nil if no error were found
func (validator *RegistrationFormValidator) Validate(form *RegistrationForm) ValidationError {
	errors := ValidationError{}
	// CSRF
	if len(strings.Trim(form.CSRF, " ")) == 0 {
		errors.Append("CSRF", "should not be empty")
	}
	if !validator.csrfProvider.Valid(form.CSRF, validator.request.RemoteAddr, "registration") {
		errors.Append("CSRF", "invalid token")
		form.CSRF = validator.csrfProvider.Generate(validator.request.RemoteAddr, "registration")
	}
	// Username
	if len(strings.Trim(form.Username, " ")) == 0 {
		errors.Append("Username", "should not be empty")
	}
	if len(form.Username) <= 5 {
		errors.Append("Username", "should be longer than 5 characters")
	}
	if len(form.Username) >= 100 {
		errors.Append("Username", "should be shorter than 100 characters")
	}
	if user, err := validator.userRepository.GetOneByUsername(form.Username); user != nil && err == nil {
		errors.Append("Username", "invalid, please choose another username")
	}
	// Email
	if len(strings.Trim(form.Email, " ")) == 0 {
		errors.Append("Email", "should not be empty")
	}
	if len(form.Email) <= 5 {
		errors.Append("Email", "should be longer than 5 characters")
	}
	if len(form.Email) >= 100 {
		errors.Append("Email", "should be shorter than 100 characters")
	}
	if !isEmail(form.Email) {
		errors.Append("Email", "invalid email")
	}
	if user, err := validator.userRepository.GetOneByEmail(form.Email); user != nil && err == nil {
		errors.Append("Email", "invalid, please choose another email")
	}
	// Password
	if len(strings.Trim(form.Password, " ")) == 0 {
		errors.Append("Password", "should not be empty")
	}
	if len(form.Password) <= 7 {
		errors.Append("Password", "should be at least 8 character long")
	}
	if len(form.Password) > 100 {
		errors.Append("Password", "should be at most 255 character long")
	}
	if form.Password != form.PasswordConfirmation {
		errors.Append("Password", "should match password confirmation")
	}
	if !errors.HasErrors() {
		return nil
	}
	form.Errors = errors
	return errors
}
