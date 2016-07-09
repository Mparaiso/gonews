package gonews

import (
	"fmt"
	"golang.org/x/net/xsrftoken"
)

// DefaultCSRFProvider implements CSRFProvider
type DefaultCSRFProvider struct {
	Session SessionInterface
	Secret  string
}

// Generate generates a new token
func (d *DefaultCSRFProvider) Generate(userID, actionID string) string {
	t := xsrftoken.Generate(d.Secret, userID, actionID)
	tokenNameInSession := fmt.Sprintf("%v-%v", userID, actionID)
	d.Session.Set(tokenNameInSession, t)
	return t
}

// Valid valides a token
func (d *DefaultCSRFProvider) Valid(token, userID, actionID string) bool {
	tokenNameInSession := fmt.Sprintf("%v-%v", userID, actionID)
	t := fmt.Sprint(d.Session.Get(tokenNameInSession))
	d.Session.Set(tokenNameInSession, nil)
	if t != token {
		return false
	}
	return xsrftoken.Valid(t, d.Secret, userID, actionID)
}
