package gonews

import (
	"bytes"

	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/gorilla/context"

	"github.com/gorilla/sessions"
)

// HandlerFunc allows http.HandlerFunc to be used as
// http.Handler
type HandlerFunc func(http.ResponseWriter, *http.Request)

func (h HandlerFunc) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	h(rw, r)
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
	if c.Debug == true {
		dump, _ := httputil.DumpRequest(r, true)
		requestDump = bytes.NewBuffer(dump).String()
	}

	c.MustGetTemplate().SetEnvironment(struct {
		FlashMessages []interface{}
		Request       string
		Description   struct{ Title string }
	}{
		c.MustGetSession(r).Flashes(),
		requestDump,
		struct{ Title string }{"GoNews"},
	})
	next()
}

// CSRFMiddleWare provides a cross site request forgergy mechanism
func CSRFMiddleWare(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	// csrfProvider := c.GetCSRFProvider()
	next()
}

// SessionMiddleware provide session capabilities
// TODO change secret key
func SessionMiddleware(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	// @see https://godoc.org/github.com/gorilla/sessions
	// for why the use of context.Clear with github.com/gorilla/sessions
	defer context.Clear(r)
	session := c.MustGetSession(r)
	// options has to be set or it will panic
	session.SetOptions(&sessions.Options{
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   60 * 60 * 24,
		Domain:   r.URL.Host,
	})

	// Set the session in the response. We need to do this because
	// we need to save the session BEFORE something is written to the http response
	rw.(ResponseWriterExtraInterface).SetSession(session)
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

// LoggingMiddleware log each request using
// comman log format
func LoggingMiddleware(c *Container, rw http.ResponseWriter, r *http.Request, next func()) {
	start := time.Now()
	next()
	// @see https://en.wikipedia.org/wiki/Common_Log_Format for log format
	// @see http://httpd.apache.org/docs/1.3/logs.html#combined
	c.MustGetLogger().Info(
		fmt.Sprintf("%s %s %s [%s] \"%s %s %s\" %s %d \"%s\" \"%s\"",
			r.RemoteAddr,
			"-",
			"-",
			start.Format("Jan/02/2006:15:04:05 -0700 MST"),
			r.Method,
			r.RequestURI,
			r.Proto,
			rw.Header().Get("Status-Code"),
			rw.(ResponseWriterExtraInterface).GetCurrentSize(),
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
