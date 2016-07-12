package gonews

import (
	"bytes"

	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/gorilla/context"
)

// AuthenticatedUserOnlyMiddleware filters out non authenticated users
func AuthenticatedUserOnlyMiddleware(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	if c.CurrentUser() == nil {
		c.MustGetSession().Delete("user.ID")
		c.MustGetSession().AddFlash(http.StatusText(http.StatusUnauthorized), "danger")
		c.ResponseWriter().WriteHeader(401)
		LoginController(c, rw, r, next)
		return
	}
	next()
}

// RefreshUserMiddleware keeps the application aware of the current user but does not authenticate or authorize
func RefreshUserMiddleware(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	session := c.MustGetSession()

	if session.Has("user.ID") {
		userID := c.MustGetSession().Get("user.ID").(int64)
		user, err := c.MustGetUserRepository().GetById(userID)
		if err == nil {
			if user != nil {
				c.SetCurrentUser(user)
			}
		} else {
			c.HTTPError(rw, r, 500, err)
			return
		}
	} else {
		c.SetCurrentUser(nil)
	}
	next()

}

//PostOnlyMiddleware filters post requests
func PostOnlyMiddleware(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	if r.Method == "POST" {
		next()
		return
	}
	c.HTTPError(rw, r, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
}

// TemplateMiddleware configure the template engine
func TemplateMiddleware(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	var requestDump string
	if c.Debug() == true {
		dump, _ := httputil.DumpRequest(r, true)
		requestDump = bytes.NewBuffer(dump).String()
	}

	c.MustGetTemplate().SetEnvironment(&TemplateEnvironment{
		FlashMessages: map[string][]interface{}{
			"error":   c.MustGetSession().Flashes("error"),
			"success": c.MustGetSession().Flashes("success"),
			"info":    c.MustGetSession().Flashes("info"),
			"notice":  c.MustGetSession().Flashes("notice"),
		},
		Request: requestDump,
		Description: struct{ Title, Slogan, Description string }{
			c.GetOptions().Title,
			c.GetOptions().Slogan,
			c.GetOptions().Description,
		},
		CurrentUser: c.CurrentUser(),
		Session:     c.MustGetSession().ValuesString(),
	})
	next()
}

// SessionMiddleware provide session capabilities
// TODO change secret key
func SessionMiddleware(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	// @see https://godoc.org/github.com/gorilla/sessions
	// for why the use of context.Clear with github.com/gorilla/sessions
	defer context.Clear(r)
	next()
}

// StopWatchMiddleware logs the request duration
func StopWatchMiddleware(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	start := time.Now()
	next()
	end := time.Now()
	duration := end.Sub(start).String()
	c.MustGetLogger().Debug(fmt.Sprintf("Request executed in %s", duration))
}

// LoggerMiddleware log each request using
// comman log format
func LoggerMiddleware(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	rw.(ResponseWriterExtra).SetLogger(c.MustGetLogger())
	start := time.Now()
	next()

	// @see https://en.wikipedia.org/wiki/Common_Log_Format for log format
	// @see http://httpd.apache.org/docs/1.3/logs.html#combined
	c.MustGetLogger().Info(
		fmt.Sprintf("%s %s %s [%s] \"%s %s %s\" %s %d \"%s\" \"%s\"",
			r.RemoteAddr,
			func() string {
				if c.CurrentUser() != nil {
					return string(c.CurrentUser().ID)
				}
				return "-"
			}(),
			func() string {
				if c.CurrentUser() != nil {
					return c.CurrentUser().Username
				}
				return "-"
			}(),
			start.Format("Jan/02/2006:15:04:05 -0700 MST"),
			r.Method,
			r.RequestURI,
			r.Proto,
			rw.Header().Get("Status-Code"),
			rw.(ResponseWriterExtra).GetCurrentSize(),
			r.Referer(),
			r.UserAgent(),
		))

}

// NotFoundMiddleware handles 404 responses
func NotFoundMiddleware(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	if r.URL.Path != "/" {
		c.HTTPError(rw, r, 404, errors.New(http.StatusText(404)))
	} else {
		next()
	}
}
