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
	"net/http"

	"fmt"

	"github.com/gorilla/sessions"
)

// SessionInterface was extracted from Session
type SessionWrapper interface {
	AddFlash(value interface{}, vars ...string)
	Flashes(vars ...string) []interface{}
	Name() string
	Save(r *http.Request, w http.ResponseWriter) error
	Store() sessions.Store
	Get(interface{}) interface{}
	Set(interface{}, interface{})
	Has(interface{}) bool
	Options() *sessions.Options
	SetOptions(*sessions.Options)
	Values() map[interface{}]interface{}
	ValuesString() map[string]interface{}
	Delete(interface{})
}

// Session implementing SessionInterface
type DefaultSessionWrapper struct {
	*sessions.Session
}

// NewSession create a new session form a store
func NewSession(store sessions.Store, request *http.Request, name string) (SessionWrapper, error) {
	session, err := store.Get(request, name)
	if err != nil {
		return nil, err
	}
	return &DefaultSessionWrapper{session}, nil
}

func (s *DefaultSessionWrapper) Options() *sessions.Options {
	return s.Session.Options
}

func (s *DefaultSessionWrapper) SetOptions(o *sessions.Options) {
	s.Session.Options = o
}

// Get gets a session value
func (s *DefaultSessionWrapper) Get(key interface{}) interface{} {
	return s.Session.Values[key]
}

// Set sets a session value
func (s *DefaultSessionWrapper) Set(key interface{}, value interface{}) {
	s.Session.Values[key] = value
}

// Has returns true if key exists
func (s *DefaultSessionWrapper) Has(key interface{}) bool {
	_, ok := s.Session.Values[key]
	return ok
}
func (s *DefaultSessionWrapper) Delete(key interface{}) {
	delete(s.Session.Values, key)
}

// Values return a map of session values
func (s *DefaultSessionWrapper) Values() map[interface{}]interface{} {
	return s.Session.Values
}

// ValuesString return a map of session values for debugging purposes
func (s *DefaultSessionWrapper) ValuesString() map[string]interface{} {
	result := map[string]interface{}{}
	for key, value := range s.Session.Values {
		result[fmt.Sprintf("%v", key)] = value
	}
	return result
}
