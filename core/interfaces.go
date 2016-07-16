package gonews

// CSRFGenerator generates and validate csrf tokens
type CSRFGenerator interface {
	Generate(userID, actionID string) string
	Valid(token, userID, actionID string) bool
}

// UserFinder can find users from a datasource
type UserFinder interface {
	GetOneByEmail(string) (*User, error)
	GetOneByUsername(string) (*User, error)
}
