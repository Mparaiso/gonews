package gonews

// CSRFProvider provide csrf tokens
type CSRFProvider interface {
	Generate(userID, actionID string) string
	Valid(token, userID, actionID string) bool
}

type UserFinder interface {
	GetOneByEmail(string) (*User, error)
	GetOneByUsername(string) (*User, error)
}
