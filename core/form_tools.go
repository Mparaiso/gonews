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
	"github.com/gorilla/Schema"
)

// FormDecoder decode url.Values into
// a struct
type FormDecoder interface {
	Decode(interface{}, map[string][]string) error
}

// DefaultFormDecoder decode url.Values into
// a struct
type DefaultFormDecoder struct {
	*schema.Decoder
}

// NewDefaultFormDecoder returns a new DefaultFormDecoder
func NewDefaultFormDecoder() *DefaultFormDecoder {
	return &DefaultFormDecoder{schema.NewDecoder()}
}

// Decode decodes values into the destination
func (d *DefaultFormDecoder) Decode(destination interface{}, values map[string][]string) error {
	return d.Decoder.Decode(destination, values)
}
