//    Gonews is a webapp that provides a forum where users can post and discuss links
//
//    Copyright (C) 2016  mparaiso <mparaiso@online.fr>
//
//    This program is free software: you can redistribute it and/or modify
//    it under the terms of the GNU Affero General Public License as published
//    by the Free Software Foundation, either version 3 of the License, or
//    (at your option) any later version.
//
//    This program is distributed in the hope that it will be useful,
//    but WITHOUT ANY WARRANTY; without even the implied warranty of
//    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//    GNU Affero General Public License for more details.
//
//    You should have received a copy of the GNU Affero General Public License
//    along with this program.  If not, see <http://www.gnu.org/licenses/>.

package gonews

import (
	// "fmt"
	"github.com/gorilla/securecookie"
	"golang.org/x/net/xsrftoken"
)

const CsrfSessionKey = "csrf-unique-id"

// CSRFGenerator generates and validate csrf tokens
type CSRFGenerator interface {
	Generate(actionID string) string
	Valid(token, actionID string) bool
}

// DefaultCSRFProvider implements CSRFProvider
type DefaultCSRFGenerator struct {
	Session SessionWrapper
	Secret  string
}

// Generate generates a new token
func (d *DefaultCSRFGenerator) Generate(actionID string) string {
	if !d.Session.Has(CsrfSessionKey) {
		d.Session.Set(CsrfSessionKey, string(securecookie.GenerateRandomKey(16)))
	}
	t := xsrftoken.Generate(d.Secret, d.Session.Get(CsrfSessionKey).(string), actionID)
	return t
}

// Valid valides a token
func (d *DefaultCSRFGenerator) Valid(token, actionID string) bool {
	return xsrftoken.Valid(token, d.Secret, d.Session.Get(CsrfSessionKey).(string), actionID)
}
