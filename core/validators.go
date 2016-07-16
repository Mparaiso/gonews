//    Gonews is a webapp that provides a forum where users can post and discuss links
//
//    Copyright (C) 2016  mparaiso <mparaiso@online.fr>

//    This program is free software: you can redistribute it and/or modify
//    it under the terms of the GNU Affero General Public License as published
//    by the Free Software Foundation, either version 3 of the License, or
//    (at your option) any later version.

//    This program is distributed in the hope that it will be useful,
//    but WITHOUT ANY WARRANTY; without even the implied warranty of
//    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//    GNU Affero General Public License for more details.

//    You should have received a copy of the GNU Affero General Public License
//    along with this program.  If not, see <http://www.gnu.org/licenses/>.

package gonews

import (
	"fmt"
	"net/http"

	"regexp"
	"strings"
)

// ValidationError is a validation error
type ValidationError interface {
	HasErrors() bool
	Append(key, value string)
	Error() string
}

// ConcreteValidationError holds errors in a map
type ConcreteValidationError map[string][]string

// Append adds an new error to a map
func (v ConcreteValidationError) Append(key string, value string) {
	v[key] = append(v[key], value)
}

func (v ConcreteValidationError) Error() string {
	return fmt.Sprintf("%#v", v)
}

// HasErrors returns true if error exists
func (v ConcreteValidationError) HasErrors() bool {
	return len(v) != 0
}

// UserValidator is a User validator
type UserValidator struct {
}

// Validate validates a user
func (uv UserValidator) Validate(u *User) ValidationError {

	errors := ConcreteValidationError{}

	StringNotEmptyValidator("Username", u.Username, &errors)
	StringMinLengthValidator("Username", u.Username, 6, &errors)
	StringMaxLengthValidator("Username", u.Username, 100, &errors)

	StringNotEmptyValidator("Email", u.Email, &errors)
	StringMinLengthValidator("Email", u.Email, 6, &errors)
	StringMaxLengthValidator("Email", u.Email, 100, &errors)
	EmailValidator("Email", u.Email, &errors)

	StringNotEmptyValidator("Password", u.Password, &errors)
	StringMinLengthValidator("Password", u.Password, 8, &errors)
	StringMaxLengthValidator("Password", u.Password, 255, &errors)

	if !errors.HasErrors() {
		return nil
	}
	return errors
}

// CommentFormValidator validates a comment form
type CommentFormValidator struct {
	CSRFGenerator
	*http.Request
}

// Validate validades a comment form
func (validator *CommentFormValidator) Validate(form *CommentForm) ValidationError {
	errors := ConcreteValidationError{}
	StringNotEmptyValidator("Content", form.Content, &errors)
	StringMinLengthValidator("Content", form.Content, 10, &errors)
	StringMaxLengthValidator("Content", form.Content, 500, &errors)

	PatternValidator("Goto", form.Goto, regexp.MustCompile(`^\/\S+\?\S+$`), &errors)
	CSRFValidator("CRSF", form.CSRF, validator.CSRFGenerator, validator.Request.RemoteAddr, "comment", &errors)
	form.CSRF = validator.CSRFGenerator.Generate(validator.Request.RemoteAddr, "comment")
	if errors.HasErrors() {
		form.Errors = errors
		return errors
	}
	return nil
}

// RegistrationFormValidator is a RegistrationForm validator
type RegistrationFormValidator struct {
	request        *http.Request
	csrfProvider   CSRFGenerator
	userRepository UserFinder
}

// NewRegistrationFormValidator creates an new RegistrationFormValidator
func NewRegistrationFormValidator(request *http.Request, csrfProvider CSRFGenerator, userFinder UserFinder) *RegistrationFormValidator {
	return &RegistrationFormValidator{request, csrfProvider, userFinder}
}

// Validate returns nil if no error were found
func (validator *RegistrationFormValidator) Validate(form *RegistrationForm) ValidationError {
	errors := ConcreteValidationError{}
	// CSRF
	StringNotEmptyValidator("CSRF", form.CSRF, &errors)
	CSRFValidator("CSRF", form.CSRF, validator.csrfProvider, validator.request.RemoteAddr, "registration", &errors)
	form.CSRF = validator.csrfProvider.Generate(validator.request.RemoteAddr, "registration")
	// Username
	StringNotEmptyValidator("Username", form.Username, &errors)
	StringMinLengthValidator("Username", form.Username, 5, &errors)
	StringMaxLengthValidator("Username", form.Username, 100, &errors)
	// validate unique username
	if user, err := validator.userRepository.GetOneByUsername(form.Username); user != nil && err == nil {
		errors.Append("Username", "invalid, please choose another username")
	}
	// Email
	StringNotEmptyValidator("Email", form.Email, &errors)
	StringMinLengthValidator("Email", form.Email, 5, &errors)
	StringMaxLengthValidator("Email", form.Email, 100, &errors)
	EmailValidator("Email", form.Email, &errors)
	// validate unique email
	if user, err := validator.userRepository.GetOneByEmail(form.Email); user != nil && err == nil {
		errors.Append("Email", "invalid, please choose another email")
	}
	// Password
	StringNotEmptyValidator("Password", form.Password, &errors)
	StringMinLengthValidator("Password", form.Password, 7, &errors)
	StringMaxLengthValidator("Password", form.Password, 255, &errors)
	MatchValidator("Password", "PasswordConfirmation", form.Password, form.PasswordConfirmation, &errors)

	if !errors.HasErrors() {
		return nil
	}
	form.Errors = errors
	return errors
}

// LoginFormValidator is a validator for LoginForm
type LoginFormValidator struct {
	csrfProvider CSRFGenerator
	request      *http.Request
}

// Validate validates a login form
func (validator *LoginFormValidator) Validate(form *LoginForm) ValidationError {
	errors := ConcreteValidationError{}
	StringNotEmptyValidator("Username", form.Username, &errors)
	StringNotEmptyValidator("Password", form.Password, &errors)
	CSRFValidator("CSRF", form.CSRF, validator.csrfProvider, validator.request.RemoteAddr, "login", &errors)
	form.CSRF = validator.csrfProvider.Generate(validator.request.RemoteAddr, "login")

	if !errors.HasErrors() {
		return nil
	}
	form.Errors = errors
	return errors
}

type SubmissionFormValidator struct {
	CSRFGenerator
	*http.Request
}

// Validate validates a submission form
func (validator *SubmissionFormValidator) Validate(form *SubmissionForm) ValidationError {
	errors := ConcreteValidationError{}

	CSRFValidator("CSRF", form.CSRF, validator.CSRFGenerator, validator.Request.RemoteAddr, "submission", &errors)
	form.CSRF = validator.CSRFGenerator.Generate(validator.Request.RemoteAddr, "submission")
	StringNotEmptyValidator("Title", form.Title, &errors)
	StringMaxLengthValidator("Title", form.Title, 100, &errors)
	StringMinLengthValidator("Title", form.Title, 5, &errors)

	switch {
	case len(strings.Trim(form.Content, " ")) == 0:
		StringNotEmptyValidator("URL", form.URL, &errors)
		StringMinLengthValidator("URL", form.URL, 5, &errors)
		StringMaxLengthValidator("URL", form.URL, 255, &errors)
		URLValidator("URL", form.URL, &errors)
	case len(strings.Trim(form.URL, " ")) == 0:
		StringNotEmptyValidator("Content", form.Content, &errors)
		StringMaxLengthValidator("Content", form.Content, 500, &errors)
		StringMinLengthValidator("Content", form.Content, 30, &errors)
	default:
		StringMinLengthValidator("URL", form.URL, 5, &errors)
		StringMaxLengthValidator("URL", form.URL, 255, &errors)
		URLValidator("URL", form.URL, &errors)

		StringMaxLengthValidator("Content", form.Content, 500, &errors)
		StringMinLengthValidator("Content", form.Content, 30, &errors)
	}

	if errors.HasErrors() {
		form.Errors = errors
		return errors
	}
	return nil
}

/*

HELPER FUNCTIONS

*/

// StringNotEmptyValidator checks if a string is empty
func StringNotEmptyValidator(field string, value string, errors ValidationError) {
	if strings.Trim(value, " ") == "" {
		errors.Append(field, "should not be empty")
	}
}

// StringMinLengthValidator validates a string by minimum length
func StringMinLengthValidator(field, value string, minlength int, errors ValidationError) {
	if len(value) < minlength {
		errors.Append(field, fmt.Sprintf("should be at least %d character long", minlength))
	}
}

// StringMaxLengthValidator validates a string by maximum length
func StringMaxLengthValidator(field, value string, maxlength int, errors ValidationError) {
	if len(value) > maxlength {
		errors.Append(field, "should be at most  %d character long")
	}
}

// MatchValidator validates a string by an expected value
func MatchValidator(field1 string, field2 string, value1, value2 interface{}, errors ValidationError) {
	if value1 != value2 {
		errors.Append(field1, fmt.Sprintf("should match %s ", field2))
	}
}

// EmailValidator validates an email
func EmailValidator(field, value string, errors ValidationError) {
	if !isEmail(value) {
		errors.Append(field, "should be a valid email")
	}
}

// URLValidator validates a URL
func URLValidator(field, value string, errors ValidationError) {
	if !IsURL(value) {
		errors.Append(field, "should be a valid URL")
	}
}

// PatternValidator valides a value according to a regexp pattern
func PatternValidator(field, value string, pattern *regexp.Regexp, errors ValidationError) {
	if !pattern.MatchString(value) {
		errors.Append(field, "should match the following pattern : "+pattern.String())
	}
}

// CSRFValidator validates a CSRF Token
func CSRFValidator(field string, value string, csrfProvider CSRFGenerator, remoteAddr, action string, errors ValidationError) {
	if !csrfProvider.Valid(value, remoteAddr, action) {
		errors.Append(field, "invalid token")
	}
}
func IsURL(candidate string) bool {
	return regexp.MustCompile(`^(https?\:\/\/)?(\w+\.)?\w+\.\w+(\.\w+)?\/?\S+$`).MatchString(candidate)
}
func isEmail(candidate string) bool {
	return regexp.MustCompile(`\w+@\w+\.\w+`).MatchString(candidate)
}
