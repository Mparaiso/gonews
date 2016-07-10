// @Copyright (c) 2016 mparaiso <mparaiso@online.fr>  All rights reserved.

package gonews

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

// ThreadIndexController displays a list of links
func ThreadIndexController(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	var threads Threads

	repository, err := c.GetThreadRepository()
	if err == nil {
		threads, err = repository.GetThreadsOrderedByVoteCount(100, 0)
		if err == nil {
			err = c.MustGetTemplate().ExecuteTemplate(rw, "thread_list.tpl.html", map[string]interface{}{
				"Threads": threads,
				"Title":   "homepage",
			})
			if err == nil {
				return
			}
		}
	}
	c.HTTPError(rw, r, 500, err)
}

// ThreadListByAuthorIDController displays user's submitted stories
func ThreadListByAuthorIDController(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 32)
	if err != nil {
		c.HTTPError(rw, r, 500, err)
	}
	userRepository := c.MustGetUserRepository()
	user, err := userRepository.GetById(id)
	if err != nil {
		c.HTTPError(rw, r, 500, err)
	}
	if user == nil {
		c.HTTPError(rw, r, 404, fmt.Sprintf("User with id %d not found", id))
	}
	threadRepository := c.MustGetThreadRepository()
	threads, err := threadRepository.GetByAuthorID(user.ID)
	if err != nil {
		c.HTTPError(rw, r, 500, err)
	}
	for _, thread := range threads {
		thread.Author = user
	}
	c.MustGetTemplate().ExecuteTemplate(rw, "user_submitted_stories.tpl.html", map[string]interface{}{"Threads": threads, "Author": user})
}

// ThreadShowController displays a thread and its comments
func ThreadShowController(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 0)
	if err != nil {
		c.HTTPError(rw, r, 500, err)
		return
	}
	repository, err := c.GetThreadRepository()
	if err != nil {
		c.HTTPError(rw, r, 500, err)
		return
	}
	thread, err := repository.GetThreadByIDWithCommentsAndTheirAuthors(int(id))
	if err != nil {
		c.HTTPError(rw, r, 500, err)
		return
	}
	if thread == nil {
		c.HTTPError(rw, r, 404, fmt.Errorf("Thread with ID %d Not Found", id))
		return
	}
	tpl, err := c.GetTemplate()
	if err != nil {
		c.HTTPError(rw, r, 500, err)
		return
	}
	err = tpl.ExecuteTemplate(rw, "thread_show.tpl.html", map[string]interface{}{"Thread": thread})
	if err != nil {
		c.HTTPError(rw, r, 500, err)
		return
	}
}

// LogoutController logs out a user
func LogoutController(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	c.MustGetSession(r).Set("user_id", nil)
	c.SetCurrentUser(nil)
	http.Redirect(rw, r, "/", http.StatusOK)
}

// LoginController displays the login/signup page
func LoginController(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	switch r.Method {
	case "GET":
		loginCSRF := c.GetCSRFProvider(r).Generate(r.RemoteAddr, "login")
		loginForm := &LoginForm{CSRF: loginCSRF, Name: "login"}
		registrationCSRF := c.GetCSRFProvider(r).Generate(r.RemoteAddr, "registration")
		registrationForm := &RegistrationForm{CSRF: registrationCSRF, Name: "registration"}
		err := c.MustGetTemplate().ExecuteTemplate(rw, "login.tpl.html", map[string]interface{}{
			"LoginForm":        loginForm,
			"RegistrationForm": registrationForm,
		})
		if err != nil {
			c.HTTPError(rw, r, 500, err)
		}
		return
	case "POST":
		c.MustGetSession(r).Set("trying to save something in session", "something")
		var loginErrorMessage string
		var candidate *User
		err := r.ParseForm()
		if err != nil {
			c.HTTPError(rw, r, 500, err)
			return
		}
		loginForm := &LoginForm{}
		err = loginForm.HandleRequest(r)
		if err != nil {
			c.HTTPError(rw, r, 500, err)
			return
		}
		loginFormValidator := &LoginFormValidator{c.GetCSRFProvider(r), r}
		err = loginFormValidator.Validate(loginForm)
		// authenticate user
		if err == nil {
			user := loginForm.Model()
			userRepository := c.MustGetUserRepository()
			candidate, err = userRepository.GetOneByUsername(user.Username)
			if err == nil && candidate != nil {
				err = candidate.Authenticate(user.Password)
				if err == nil {
					// authenticated
					c.MustGetSession(r).Set("user.ID", candidate.ID)
					c.MustGetSession(r).Set("trying to save something in session", "something")
					c.MustGetLogger().Debug("auth sucessful, redirecting")
					http.Redirect(rw, r, "/", 301)
					return
				}
			} else if candidate == nil {
				loginErrorMessage = "Invalid Credentials"
			}
		}

		rw.WriteHeader(http.StatusBadRequest)
		registrationCSRF := c.GetCSRFProvider(r).Generate(r.RemoteAddr, "registration")
		registrationForm := &RegistrationForm{CSRF: registrationCSRF, Name: "registration"}
		c.MustGetLogger().Error(err)
		err = c.MustGetTemplate().ExecuteTemplate(rw, "login.tpl.html", map[string]interface{}{
			"LoginForm":         loginForm,
			"RegistrationForm":  registrationForm,
			"LoginErrorMessage": loginErrorMessage,
		})
		return

	default:
		c.HTTPError(rw, r, http.StatusNotFound, http.StatusText(http.StatusNotFound))
	}
}

// RegistrationController handles user registration
func RegistrationController(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	// Parse form
	err := r.ParseForm()
	if err != nil {
		c.HTTPError(rw, r, 500, err)
		return
	}
	registrationForm := &RegistrationForm{}
	err = registrationForm.HandleRequest(r)
	if err != nil {
		c.HTTPError(rw, r, 500, err)
		return
	}
	registrationFormValidator := NewRegistrationFormValidator(r, c.GetCSRFProvider(r), c.MustGetUserRepository())
	validationError := registrationFormValidator.Validate(registrationForm)
	if validationError != nil {
		c.MustGetLogger().Error(validationError)
		c.MustGetSession(r).AddFlash("Registration Form has errors", "errors")
		rw.WriteHeader(http.StatusBadRequest)
		tErr := c.MustGetTemplate().ExecuteTemplate(rw, "login.tpl.html", map[string]interface{}{
			"RegistrationForm": registrationForm,
		})
		c.MustGetLogger().Error(tErr)
		return
	}
	user := registrationForm.Model()
	user.CreateSecurePassword(user.Password)
	err = c.MustGetUserRepository().Save(user)
	if err != nil {
		c.HTTPError(rw, r, http.StatusInternalServerError, err)
		return
	}
	c.MustGetSession(r).AddFlash("Registration Successful, please login", "success")
	http.Redirect(rw, r, "/login", http.StatusCreated)
}

// UserShowController displays the user's informations
func UserShowController(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 0)
	if err != nil {
		c.HTTPError(rw, r, 500, err)
		return
	}
	user, err := c.MustGetUserRepository().GetById(id)
	if err != nil {
		c.HTTPError(rw, r, 500, err)
		return
	}
	if user == nil {
		c.HTTPError(rw, r, 404, errors.New(http.StatusText(404)))
		return
	}
	err = c.MustGetTemplate().ExecuteTemplate(rw, "user_profile.tpl.html", map[string]interface{}{"User": user})
	if err != nil {
		c.HTTPError(rw, r, 500, err)
	}
}

// NotFoundController is a standard 404 page
func NotFoundController(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	c.HTTPError(rw, r, 404, errors.New(http.StatusText(404)))
}
