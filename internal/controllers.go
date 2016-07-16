// @Copyright (c) 2016 mparaiso <mparaiso@online.fr>  All rights reserved.

package gonews

import (
	"database/sql"
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

// ThreadByHostController displays a list of threads sharing the same host
func ThreadByHostController(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	host := c.Request().URL.Query().Get("site")
	threads, err := c.MustGetThreadRepository().GetWhereURLLike("%" + host + "%")
	if err == nil {
		err = c.MustGetTemplate().ExecuteTemplate(rw, "thread_list.tpl.html", map[string]interface{}{
			"Threads": threads,
			"Title":   "Stories by domain " + host,
		})
	}
	if err != nil {
		c.HTTPError(rw, r, http.StatusInternalServerError, err)
	}
}

// CommentsByAuthorController displays comments by author
func CommentsByAuthorController(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	id, err := strconv.ParseInt(c.Request().URL.Query().Get("id"), 10, 64)
	if err != nil {
		c.HTTPError(rw, r, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}
	var (
		author   *User
		comments Comments
	)
	author, err = c.MustGetUserRepository().GetByID(id)
	if err == nil {
		comments, err = c.MustGetCommentRepository().GetCommentsByAuthorID(id)
		if err == nil {
			err = c.MustGetTemplate().ExecuteTemplate(rw, "comments_list.tpl.html", map[string]interface{}{
				"Comments": comments,
				"Author":   author,
				"Title":    fmt.Sprintf("%s's comments", author.Username),
			})
		}
	}
	if err != nil {
		c.HTTPError(rw, r, http.StatusInternalServerError, err)
	}
}

// ThreadListByAuthorIDController displays user's submitted stories
func ThreadListByAuthorIDController(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 32)
	if err != nil {
		c.HTTPError(rw, r, 500, err)
		return
	}
	userRepository := c.MustGetUserRepository()
	user, err := userRepository.GetByID(id)
	if err != nil {
		c.HTTPError(rw, r, 500, err)
		return
	}
	if user == nil {
		c.HTTPError(rw, r, http.StatusNotFound, fmt.Sprintf("User with id %d not found", id))
		return
	}
	threadRepository := c.MustGetThreadRepository()
	threads, err := threadRepository.GetByAuthorID(user.ID)
	if err == sql.ErrNoRows {
		c.HTTPError(rw, r, http.StatusNotFound, err)
		return
	} else if err != nil {
		c.HTTPError(rw, r, http.StatusInternalServerError, err)
		return
	}
	for _, thread := range threads {
		thread.Author = user
	}
	err = c.MustGetTemplate().ExecuteTemplate(rw, "user_submitted_stories.tpl.html", map[string]interface{}{"Threads": threads, "Author": user})
	if err != nil {
		c.HTTPError(rw, r, http.StatusInternalServerError, err)
	}
}

// ThreadShowController displays a thread and its comments
func ThreadShowController(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 0)
	if err != nil {
		c.HTTPError(rw, r, 500, err)
		return
	}

	thread, err := c.MustGetThreadRepository().GetThreadByIDWithCommentsAndTheirAuthors(int(id))
	if err != nil {
		c.HTTPError(rw, r, 500, err)
		return
	}
	if thread == nil {
		c.HTTPError(rw, r, 404, fmt.Errorf("Thread with ID %d Not Found", id))
		return
	}
	comment := &Comment{ThreadID: thread.ID, ParentID: 0}
	if c.HasAuthenticatedUser() {
		comment.AuthorID = c.CurrentUser().ID
	}
	commentForm := &CommentForm{Goto: fmt.Sprintf("/item?id=%d", id)}
	commentForm.SetModel(comment)
	err = c.MustGetTemplate().ExecuteTemplate(rw, "thread_show.tpl.html", map[string]interface{}{
		"Thread":      thread,
		"CommentForm": commentForm,
	})
	if err != nil {
		c.HTTPError(rw, r, 500, err)
		return
	}
}

// LogoutController logs out a user
func LogoutController(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	c.MustGetSession().Delete("user.ID")
	c.HTTPRedirect("/", 302)
}

// LoginController displays the login/signup page
func LoginController(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	switch r.Method {
	case "GET":
		loginCSRF := c.MustGetCSRFGenerator().Generate(r.RemoteAddr, "login")
		loginForm := &LoginForm{CSRF: loginCSRF, Name: "login"}
		registrationCSRF := c.MustGetCSRFGenerator().Generate(r.RemoteAddr, "registration")
		registrationForm := &RegistrationForm{CSRF: registrationCSRF, Name: "registration"}
		err := c.MustGetTemplate().ExecuteTemplate(rw, "login.tpl.html", map[string]interface{}{
			"LoginForm":        loginForm,
			"RegistrationForm": registrationForm,
		})
		if err != nil {
			c.HTTPError(rw, r, http.StatusInternalServerError, err)
		}
		return
	case "POST":
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
		loginFormValidator := &LoginFormValidator{c.MustGetCSRFGenerator(), r}
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
					c.MustGetSession().Set("user.ID", candidate.ID)
					c.HTTPRedirect("/", 302)
					return
				}
			} else if candidate == nil {
				loginErrorMessage = "Invalid Credentials"
			}
		}

		rw.WriteHeader(http.StatusBadRequest)
		registrationCSRF := c.MustGetCSRFGenerator().Generate(r.RemoteAddr, "registration")
		registrationForm := &RegistrationForm{CSRF: registrationCSRF, Name: "registration"}
		c.MustGetLogger().Error(err)
		err = c.MustGetTemplate().ExecuteTemplate(rw, "login.tpl.html", map[string]interface{}{
			"LoginForm":         loginForm,
			"RegistrationForm":  registrationForm,
			"LoginErrorMessage": loginErrorMessage,
		})
		if err != nil {
			c.HTTPError(rw, r, http.StatusInternalServerError, err)
		}
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
	registrationFormValidator := NewRegistrationFormValidator(r, c.MustGetCSRFGenerator(), c.MustGetUserRepository())
	validationError := registrationFormValidator.Validate(registrationForm)
	if validationError != nil {
		c.MustGetLogger().Error(validationError)
		c.MustGetSession().AddFlash("Registration Form has errors", "errors")
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
	c.MustGetSession().AddFlash("Registration Successful, please login", "success")
	c.HTTPRedirect("/login", 302)
}

// UserShowController displays the user's informations
func UserShowController(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 0)
	if err != nil {
		c.HTTPError(rw, r, 500, err)
		return
	}
	user, err := c.MustGetUserRepository().GetByID(id)
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
		c.HTTPError(rw, r, http.StatusInternalServerError, err)
	}
}

// SubmissionController handles submitted stories
func SubmissionController(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	user := c.CurrentUser()
	if user == nil {
		c.HTTPRedirect("/login", http.StatusUnauthorized)
		return
	}
	thread := &Thread{}
	submissionForm := &SubmissionForm{CSRF: c.MustGetCSRFGenerator().Generate(r.RemoteAddr, "submission")}
	submissionForm.SetModel(thread)
	switch r.Method {
	case "GET":
		err := c.MustGetTemplate().ExecuteTemplate(rw, "submit.tpl.html", map[string]interface{}{
			"SubmissionForm": submissionForm,
		})
		if err != nil {
			c.MustGetLogger().Error(err)
		}
	case "POST":
		err := submissionForm.HandleRequest(r)
		if err != nil {
			c.HTTPError(rw, r, 500, err)
			return
		}
		submissionFormValidator := &SubmissionFormValidator{c.MustGetCSRFGenerator(), r}
		err = submissionFormValidator.Validate(submissionForm)
		if err == nil {
			thread := submissionForm.Model()
			thread.AuthorID = user.ID
			err = c.MustGetThreadRepository().Create(thread)
			if err == nil {
				c.MustGetSession().AddFlash("Story successfully created!", "success")
				c.HTTPRedirect(fmt.Sprintf("/item?id=%d", thread.ID), 302)
				return
			}
		}
		c.ResponseWriter().WriteHeader(http.StatusBadRequest)
		c.MustGetLogger().Error(err)
		err = c.MustGetTemplate().ExecuteTemplate(rw, "submit.tpl.html", map[string]interface{}{
			"SubmissionForm": submissionForm,
		})
		if err != nil {
			c.HTTPError(rw, r, http.StatusInternalServerError, err)
		}
	}

}

// CommentCreateController handles comment submission
func CommentCreateController(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	switch r.Method {
	case "GET":
		// reply
		type Query struct {
			ID   int64
			Goto string
		}
		q := new(Query)
		err := decoder.Decode(q, r.URL.Query()) /// TODO find out why the following doesn't work c.GetFormDecoder().Decode(&q, r.URL.Query())
		if err != nil {
			c.HTTPError(rw, r, http.StatusBadRequest, err)
			return
		}
		parentComment, err := c.MustGetCommentRepository().GetByID(q.ID)
		if err != nil {
			c.HTTPError(rw, r, http.StatusInternalServerError, err)
			return
		}
		if parentComment == nil {
			c.HTTPError(rw, r, http.StatusNotFound, http.StatusText(http.StatusNotFound))
			return
		}
		comment := &Comment{ThreadID: parentComment.ThreadID, ParentID: parentComment.ID}

		form := &CommentForm{CSRF: c.MustGetCSRFGenerator().Generate(r.RemoteAddr, "comment"), Goto: q.Goto}
		form.SetModel(comment)
		err = c.MustGetTemplate().ExecuteTemplate(rw, "comment_create.tpl.html", map[string]interface{}{
			"ParentComment": parentComment,
			"CommentForm":   form,
			"Title":         "Reply",
		})
		if err != nil {
			c.HTTPError(rw, r, http.StatusInternalServerError, err)
		}
		return
	case "POST":
		form := &CommentForm{CSRF: c.MustGetCSRFGenerator().Generate(r.RemoteAddr, "comment")}
		form.SetModel(&Comment{AuthorID: c.CurrentUser().ID})
		err := form.HandleRequest(r)
		if err != nil {
			c.HTTPError(rw, r, 500, err)
			return
		}
		formValidator := &CommentFormValidator{c.MustGetCSRFGenerator(), r}
		err = formValidator.Validate(form)
		if err == nil {
			comment := form.Model()
			err = c.MustGetCommentRepository().Create(comment)
			if err == nil {
				c.MustGetSession().AddFlash("Comment sucessfully created.", "success")
				c.HTTPRedirect(form.Goto, 302)
				return
			}
		}
		c.ResponseWriter().WriteHeader(http.StatusBadRequest)
		err = c.MustGetTemplate().ExecuteTemplate(rw, "comment_form.tpl.html", map[string]interface{}{
			"CommentForm": form,
			"Title":       "CommentForm",
			"Error":       "Your form has errors",
		})
		if err != nil {
			c.HTTPError(rw, r, 500, err)
		}
	default:
		c.HTTPError(rw, r, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
	}
}

// NewCommentsController displays new comments
func NewCommentsController(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	var (
		err      error
		comments Comments
	)
	comments, err = c.MustGetCommentRepository().GetNewestComments()
	if err == nil {
		err = c.MustGetTemplate().ExecuteTemplate(rw, "newcomments.tpl.html", map[string]Any{
			"Title":    "New Comments",
			"Comments": comments,
		})
	}
	if err != nil {
		c.HTTPError(rw, r, 500, err)
	}
}

// NewestStoriesController displays new stories
func NewestStoriesController(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	var (
		stories Threads
		err     error
	)
	stories, err = c.MustGetThreadRepository().GetNewest()
	if err == nil {
		err = c.MustGetTemplate().ExecuteTemplate(rw, "thread_list.tpl.html", map[string]interface{}{
			"Title":   "New Stories",
			"Threads": stories,
		})
	}
	if err != nil {
		c.HTTPError(rw, r, 500, err)
	}
}

// NotFoundController is a standard 404 page
func NotFoundController(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	c.HTTPError(rw, r, http.StatusNotFound, http.StatusText(http.StatusNotFound))
}
