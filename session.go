package gonews

import (
	"net/http"

	"github.com/gorilla/sessions"
)

// SessionInterface was extracted from Session
type SessionInterface interface {
	AddFlash(value interface{}, vars ...string)
	Flashes(vars ...string) []interface{}
	Name() string
	Save(r *http.Request, w http.ResponseWriter) error
	Store() sessions.Store
	Get(Any) Any
	Set(Any, Any)
	Has(Any) bool
	Options() *sessions.Options
	SetOptions(*sessions.Options)
}

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

func (s *Session) Has(key Any) bool {
	_, ok := s.Session.Values[key]
	return ok
}
