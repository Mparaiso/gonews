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
	"html/template"
	"io"
)

// TemplateEnvironment is used to store
// global data common to all templates .
// it is available as .Environment variable in all templates .
// Data specific to a controller are available through .Data variable .
type TemplateEnvironment struct {
	FlashMessages map[string][]interface{}
	Request       string
	Description   struct{ Title, Slogan, Description string }
	Configuration struct{ CommentMaxDepth int }
	CurrentUser   *User
	Session       map[string]interface{}
}

// TemplateProvider provides templates
type TemplateEngine interface {
	ExecuteTemplate(io.Writer, string, interface{}) error
	Environment() Any
	SetEnvironment(Any)
}

// Template implement template provider
type DefaultTemplateEngine struct {
	*template.Template
	environment Any
}

// Environment returns the environement used in
// templates. then Environment is passed to every template
// being rendered
func (t *DefaultTemplateEngine) Environment() Any {
	return t.environment
}

// SetEnvironment sets the environment passed to every rendered template
func (t *DefaultTemplateEngine) SetEnvironment(env Any) {
	t.environment = env
}

// ExecuteTemplate renders a template
func (t *DefaultTemplateEngine) ExecuteTemplate(writer io.Writer, name string, data interface{}) error {
	// We need to use a temporary buffer.
	// The reason is that ExecuteTemplate may return an error,
	// We want to be able to catch it and return status 500 if needed
	templateBuffer := new(bytes.Buffer)
	err := t.Template.ExecuteTemplate(templateBuffer, name, struct {
		Data        Any
		Environment Any
	}{data, t.environment})
	if err == nil {
		_, err = templateBuffer.WriteTo(writer)
	}
	return err
}
