package gonews

import (
	"net/http"

	"fmt"
	"github.com/gorilla/sessions"
)

type Session struct {
	*sessions.Session
}

// NewSession create a new session form a store
func NewSession(store sessions.Store, request *http.Request, name string) (SessionInterface, error) {
	session, err := store.Get(request, name)
	if err != nil {
		return nil, err
	}
	return &Session{session}, nil
}

func (s *Session) Options() *sessions.Options {
	return s.Session.Options
}

func (s *Session) SetOptions(o *sessions.Options) {
	s.Session.Options = o
}

func (s *Session) Get(key Any) Any {
	return s.Session.Values[key]
}

func (s *Session) Set(key Any, value Any) {
	s.Session.Values[key] = value
}

// Has returns true if key exists
func (s *Session) Has(key Any) bool {
	_, ok := s.Session.Values[key]
	return ok
}

// Values return a map of session values
func (s *Session) Values() map[string]interface{} {
	result := map[string]interface{}{}
	for key, value := range s.Session.Values {
		result[fmt.Sprintf("%v", key)] = value
	}
	return result
}
