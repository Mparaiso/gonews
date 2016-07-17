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
		user, err := c.MustGetUserRepository().GetByID(userID)
		if err == nil {
			if user != nil {
				c.SetCurrentUser(user)
			} else {
				session.Delete("user.ID")
				c.SetCurrentUser(nil)
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
		Configuration: struct {
			CommentMaxDepth int
		}{
			CommentMaxDepth: c.GetOptions().CommentMaxDepth,
		},
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
		fmt.Sprintf("%s %s %s [%s] \"%s %s %s\" %d %d \"%s\" \"%s\"",
			r.RemoteAddr,
			func() string {
				if c.CurrentUser() != nil {
					return fmt.Sprintf("%d", c.CurrentUser().ID)
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
			c.ResponseWriter().Status(),
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

// FaviconMiddleware returns an empty response when favicon.ico requested
func FaviconMiddleware(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	if r.URL.Path == "/favicon.ico" {
		rw.Write([]byte{0})
		return
	}
	next()
}
