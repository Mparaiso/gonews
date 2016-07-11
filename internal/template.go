package gonews

import (
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
	CurrentUser   *User
	Session       map[string]interface{}
}

// TemplateProvider provides templates
type TemplateProvider interface {
	ExecuteTemplate(io.Writer, string, interface{}) error
	Environment() Any
	SetEnvironment(Any)
}

// Template implement template provider
type Template struct {
	*template.Template
	environment Any
}

// Environment returns the environement used in
// templates. then Environment is passed to every template
// being rendered
func (t *Template) Environment() Any {
	return t.environment
}

// SetEnvironment sets the environment passed to every rendered template
func (t *Template) SetEnvironment(env Any) {
	t.environment = env
}

// ExecuteTemplate renders a template
func (t *Template) ExecuteTemplate(writer io.Writer, name string, data interface{}) error {
	return t.Template.ExecuteTemplate(writer, name, struct {
		Data        Any
		Environment Any
	}{data, t.environment})
}
