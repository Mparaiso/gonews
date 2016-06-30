package hn

import (
	"golang.org/x/crypto/bcrypt"
	"time"
)

// User is a forum user
type User struct {
	ID       int
	Username string
	Password string
	Email    string
	Created  time.Time
	Updated  time.Time
}

// CreateSecurePassword generates a secure password from a string
// and sets User.Password
func (u *User) CreateSecurePassword(unecryptedpassword string) error {
	password, err := bcrypt.GenerateFromPassword([]byte(unecryptedpassword), 0)
	if err != nil {
		return err
	}
	u.Password = string(password)
	return nil
}

// Authenticate matches unecryptedpassword with User.Password
// returns an error if they do not match
func (u *User) Authenticate(unecryptedpassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(unecryptedpassword))
	if err != nil {
		return err
	}
	return nil
}

// UserRole is a relation between users and roles
type UserRole struct {
	ID     int
	UserID int
	RoleID int
}

// Role is a role
type Role struct {
	ID   int
	Name string
}

// Thread is a forum thread
type Thread struct {
	ID      int
	Title   string
	URL     string
	Created time.Time
	Updated time.Time
}

// Comment is a comment in a thread
type Comment struct {
	ID       int
	ParentID int
	AuthorID int
	ThreadID int
	Content  string
	Created  time.Time
	Updated  time.Time
}

// Comment vote is a vote for a comment
type CommentVote struct {
	ID        int
	CommentID int
	AuthorID  int
	Value     int
	Created   time.Time
	Updated   time.Time
}

// ThreadVote is a vote for a thread
type ThreadVote struct {
	ID       int
	ThreadID int
	AuthorID int
	Value    int
	Created  time.Time
	Updated  time.Time
}
