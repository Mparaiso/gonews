package main_test

import (
	"testing"

	"database/sql"
	"github.com/mparaiso/go-news/internal"
	"net/http"
)

func Test_UserValidator_Validate_valid_user(t *testing.T) {
	user := &gonews.User{Username: "Bill_Doe", Email: "bill.doe@acme.com"}
	user.MustCreateSecurePassword("password")
	userValidator := &gonews.UserValidator{}
	if err := userValidator.Validate(user); err != nil {
		t.Fatalf("user should be valid, got '%v'", err)
	}
}

func Test_UserValidator_Validate_invalid_user(t *testing.T) {
	user := &gonews.User{Username: "Bill_Doe", Email: "bill.doe@acme"}
	user.MustCreateSecurePassword("password")
	userValidator := &gonews.UserValidator{}
	if err := userValidator.Validate(user); err == nil {
		t.Fatalf("user should be invalid, got '%v'", err)
	}
}

func Test_RegistrationFormValidator_valid_registrationForm(t *testing.T) {
	form := &gonews.RegistrationForm{CSRF: "csrf-token", Username: "johnny_doe", Password: "password", PasswordConfirmation: "password", Email: "johnny_doe@acme.com"}
	r := new(http.Request)
	r.RemoteAddr = "some-addr"
	validator := gonews.NewRegistrationFormValidator(r, TestCSRFProvider{}, TestUserFinder{})
	errors := validator.Validate(form)
	if errors != nil {
		t.Fatal("There should be no error got : ", errors)
	}
}

// CSRFProvider provide csrf tokens
type TestCSRFProvider struct{}

func (TestCSRFProvider) Generate(userID, actionID string) string {
	return "csrf-token"
}
func (TestCSRFProvider) Valid(token, userID, actionID string) bool {
	return token == "csrf-token"
}

type TestUserFinder struct{}

func (TestUserFinder) GetOneByEmail(email string) (*gonews.User, error) {
	if email == "john_doe@acme.com" {
		return &gonews.User{}, nil
	}
	return nil, sql.ErrNoRows
}
func (TestUserFinder) GetOneByUsername(username string) (*gonews.User, error) {
	if username == "john_doe" {
		return &gonews.User{}, nil
	}
	return nil, sql.ErrNoRows
}
