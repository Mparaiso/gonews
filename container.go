// @Copyright (c) 2016 mparaiso <mparaiso@online.fr>  All rights reserved.

package gonews

import (
	"bytes"
	"database/sql"
	"encoding/json"

	"html/template"
	"log"
	"net/http"
	"os"

	"fmt"

	"errors"

	"github.com/gorilla/sessions"
)

// Any is any value
type Any interface{}

// ContainerOptions are options provided to the container
type ContainerOptions struct {
	DataSource,
	Driver,
	Secret,
	TemplateDirectory string
	Debug               bool
	SessionStoreFactory func() (sessions.Store, error)
	ConnectionFactory   func() (*sql.DB, error)
	LoggerFactory       func() (LoggerInterface, error)
	csrfProvider        CSRFProvider
}

// Container contains all the application dependencies
type Container struct {
	ContainerOptions
	db               *sql.DB
	logger           LoggerInterface
	threadRepository *ThreadRepository
	userRepository   *UserRepository
	template         TemplateProvider
	session          SessionInterface
}

func (c *Container) GetSecret() string {
	return c.ContainerOptions.Secret
}

// GetTemplate returns *template.Template
func (c *Container) GetTemplate() (TemplateProvider, error) {
	if c.template == nil {
		tpl, err := template.New("templates").Funcs(template.FuncMap{
			"IsDebug": func() bool {
				return c.Debug
			},
			"ToJson": func(object Any) (string, error) {
				b, err := json.MarshalIndent(object, "", "\t")
				if err != nil {
					return "", err
				}
				return bytes.NewBuffer(b).String(), err
			},
		}).ParseGlob(c.ContainerOptions.TemplateDirectory + "/*.tpl.html")

		if err != nil {
			return nil, err
		}
		c.template = &Template{Template: tpl}
	}
	return c.template, nil
}

// MustGetTemplate panics on error
func (c *Container) MustGetTemplate() TemplateProvider {
	tpl, err := c.GetTemplate()
	if err != nil {
		panic(err)
	}
	return tpl
}

// GetConnection returns *sql.DB
func (c *Container) GetConnection() (*sql.DB, error) {
	if c.ContainerOptions.ConnectionFactory != nil {
		db, err := c.ContainerOptions.ConnectionFactory()
		if err != nil {
			return nil, err
		}
		c.db = db
	} else if c.db == nil {
		db, err := sql.Open(c.ContainerOptions.Driver, c.ContainerOptions.DataSource)
		if err != nil {
			return nil, err
		}
		c.db = db
	}
	return c.db, nil
}

// GetThreadRepository returns a repository for Thread
func (c *Container) GetThreadRepository() (*ThreadRepository, error) {
	if c.threadRepository == nil {
		db, err := c.GetConnection()
		if err != nil {
			return nil, err
		}
		c.threadRepository = &ThreadRepository{DB: db, Logger: c.MustGetLogger()}
	}
	return c.threadRepository, nil
}

func (c *Container) MustGetThreadRepository() *ThreadRepository {
	r, err := c.GetThreadRepository()
	if err != nil {
		panic(err)
	}
	return r
}

// GetUserRepository returns a repository for User
func (c *Container) GetUserRepository() (*UserRepository, error) {
	if c.userRepository == nil {
		db, err := c.GetConnection()
		if err != nil {
			return nil, err
		}
		logger, err := c.GetLogger()
		if err != nil {
			return nil, err
		}
		c.userRepository = &UserRepository{db, logger}
	}
	return c.userRepository, nil
}

// MustGetUserRepository panics on error or return a repository of User
func (c *Container) MustGetUserRepository() *UserRepository {
	r, err := c.GetUserRepository()
	if err != nil {
		panic(err)
	}
	return r
}

func (c *Container) GetCSRFProvider(request *http.Request) CSRFProvider {
	if c.csrfProvider == nil {
		c.csrfProvider = &DefaultCSRFProvider{c.MustGetSession(request), c.GetSecret()}
	}
	return c.csrfProvider
}

// GetLogger gets a logger
func (c *Container) GetLogger() (LoggerInterface, error) {
	if c.logger == nil {
		if c.ContainerOptions.LoggerFactory != nil {
			logger, err := c.ContainerOptions.LoggerFactory()
			if err != nil {
				return nil, err
			}
			c.logger = logger
		} else {
			logger := &log.Logger{}
			logger.SetOutput(os.Stdout)
			c.logger = &Logger{logger, c.Debug}
		}
	}
	return c.logger, nil
}

// MustGetLogger panics on error or return a LoggerInterface
func (c *Container) MustGetLogger() LoggerInterface {
	logger, err := c.GetLogger()
	if err != nil {
		panic(err)
	}
	return logger
}

// HTTPError writes an error to the response
func (c *Container) HTTPError(rw http.ResponseWriter, r *http.Request, status int, message Any) {
	c.MustGetLogger().Error(fmt.Sprintf("%s %d %s", r.URL, status, message))
	rw.WriteHeader(status)
	// if debug show a detailed error message
	if c.Debug == true {
		// if response has been sent, just write to output for now
		// TODO buffer response in order to handle the case where there is
		// 		an error in the template which should lead to a status 500
		if rw.(ResponseWriterExtraInterface).IsResponseWritten() {
			http.Error(rw, fmt.Sprintf("%v", message), status)
			return
		}
		// if not then execute the template with the Message
		c.MustGetTemplate().ExecuteTemplate(rw, "error.tpl.html", map[string]interface{}{"Status": status, "Message": message})
		return
	}
	// if not debug show a generic error message.
	// don't show a detailed error message
	if rw.(ResponseWriterExtraInterface).IsResponseWritten() {
		http.Error(rw, http.StatusText(status), status)
		return
	}
	c.MustGetTemplate().ExecuteTemplate(rw, "error.tpl.html", map[string]interface{}{"Status": status, "Message": http.StatusText(status)})
}

// GetSessionStore returns a session.Store
func (c *Container) GetSessionStore() (sessions.Store, error) {
	if c.ContainerOptions.SessionStoreFactory == nil {
		return nil, errors.New("SessionStoreFactory not defined in Container.Options")
	}
	return c.ContainerOptions.SessionStoreFactory()
}

// GetSession returns the session
func (c *Container) GetSession(request *http.Request) (SessionInterface, error) {
	if c.session == nil {
		sessionStore, err := c.GetSessionStore()
		if err != nil {
			return nil, err
		}
		session, err := NewSession(sessionStore, request, "gonews-session")
		if err != nil {
			return nil, err
		}
		c.session = session
	}
	return c.session, nil
}

// MustGetSession panics on error
func (c *Container) MustGetSession(request *http.Request) SessionInterface {
	session, err := c.GetSession(request)
	if err != nil {
		panic(err)
	}
	return session
}
