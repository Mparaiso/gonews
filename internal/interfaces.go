package gonews

import (
	"net/http"
)

// Form interface is a form
type Form interface {
	// HandleRequest deserialize the request body into a form struct
	HandleRequest(r *http.Request) error
}

// CSRFProvider provide csrf tokens
type CSRFProvider interface {
	Generate(userID, actionID string) string
	Valid(token, userID, actionID string) bool
}

// UserFinder can find users from a datasource
type UserFinder interface {
	GetOneByEmail(string) (*User, error)
	GetOneByUsername(string) (*User, error)
}

// ValidationError is a validation error
type ValidationError interface {
	HasErrors() bool
	Append(key, value string)
	Error() string
}
