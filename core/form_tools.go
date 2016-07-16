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
