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
	Submit               string `schema:"registration_submit"`
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

// LoginForm implements Form
type LoginForm struct {
	Name     string
	CSRF     string `schema:"login_csrf"`
	Username string `schema:"login_username"`
	Password string `schema:"login_password"`
	Submit   string `schema:"login_submit"`
	Errors   map[string][]string
	model    *User
}

// HandleRequest deserialize the request body into a form struct
func (form *LoginForm) HandleRequest(r *http.Request) error {
	return decoder.Decode(form, r.PostForm)
}

// Model return the underlying form model
func (form *LoginForm) Model() *User {
	if form.model == nil {
		form.model = &User{
			Username: form.Username,
			Password: form.Password,
		}
	}
	return form.model
}

// SubmissionForm is a submission form
type SubmissionForm struct {
	Name    string
	CSRF    string `schema:"submission_csrf"`
	Title   string `schema:"submission_title"`
	URL     string `schema:"submission_url"`
	Content string `schema:"submission_content"`
	Submit  string `schema:"submission_submit"`
	Errors  map[string][]string
	model   *Thread
}

// HandleRequest deserialize the request body into a form struct
func (form *SubmissionForm) HandleRequest(r *http.Request) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}
	return decoder.Decode(form, r.PostForm)
}

func (form *SubmissionForm) SetModel(thread *Thread) {
	form.model = thread
	form.Content = thread.Content
	form.URL = thread.URL
	form.Title = thread.Title
}

// Model return the underlying form model
func (form *SubmissionForm) Model() *Thread {
	if form.model != nil {
		form.model.Title = form.Title
		form.model.Content = form.Content
		form.model.URL = form.URL
	}
	return form.model
}
