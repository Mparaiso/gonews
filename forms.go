// @Copyright (c) 2016 mparaiso <mparaiso@online.fr>  All rights reserved.

package gonews

import (
	"net/http"

	"github.com/gorilla/Schema"
)

var decoder = schema.NewDecoder()

// RegistrationForm is a registration form
type RegistrationForm struct {
	Name                 string
	CSRF                 string `schema:"registration_csrf"`
	Username             string `schema:"registration_username"`
	Password             string `schema:"registration_password"`
	PasswordConfirmation string `schema:"registration_password_confirmation"`
	Email                string `schema:"registration_email"`
	Errors               map[string][]string
}

func (form *RegistrationForm) Model() *User {
	return &User{
		Username: form.Username,
		Password: form.Password,
		Email:    form.Email,
	}
}

// HandleRequest populates form values from request or return an error
// if it can't populate the form
func (form *RegistrationForm) HandleRequest(r *http.Request) error {
	return decoder.Decode(form, r.PostForm)
}

// LoginForm is a login form
type LoginForm struct {
	Name     string
	CSRF     string
	Username string
	Password string
	Errors   map[string][]string
}
