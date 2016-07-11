package gonews

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User is a forum user
type User struct {
	ID       int64
	Username string
	Password string
	Email    string

	Created time.Time
	Updated time.Time
	// Virtual
	Karma int
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

// MustCreateSecurePassword can panic
func (u *User) MustCreateSecurePassword(unecryptedpassword string) {
	err := u.CreateSecurePassword(unecryptedpassword)
	if err != nil {
		panic(err)
	}
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
	ID     int64
	UserID int
	RoleID int
}

// Role is a role
type Role struct {
	ID   int64
	Name string
}

// Thread is a forum thread
type Thread struct {
	ID      int64
	Title   string
	URL     string
	Content string
	Created time.Time
	Updated time.Time
	// Score is a virtual column, counts all the votes for a thread
	Score int
	// CommentCount
	CommentCount int
	// Author is the author of the thread
	Author   *User
	AuthorID int64
	// Comments in the thread
	Comments Comments
}

// Comment is a comment in a thread
type Comment struct {
	ID           int64
	ParentID     int64
	AuthorID     int64
	AuthorName   string
	ThreadID     int64
	Content      string
	CommentScore int
	Created      time.Time
	Updated      time.Time
	Children     Comments
}

func (c *Comment) HasChildren() bool {
	return len(c.Children) > 0
}

type Comments []*Comment

func (c Comments) Len() int { return len(c) }

func (c Comments) At(index int) *Comment {
	if len(c) <= (index + 1) {
		return c[index]
	}
	return nil
}

// GetTree Builds a tree of comments
func (c Comments) GetTree() (commentTree []*Comment) {
	for _, comment := range c {
		id := comment.ID
		if comment.ParentID == 0 {
			commentTree = append(commentTree, comment)
		}
		for _, subComment := range c {
			if subComment.ParentID == id {
				comment.Children = append(comment.Children, subComment)
			}
		}
	}
	return
}

// Comment vote is a vote for a comment
type CommentVote struct {
	ID        int64
	CommentID int64
	AuthorID  int64
	Value     int
	Created   time.Time
	Updated   time.Time
}

// ThreadVote is a vote for a thread
type ThreadVote struct {
	ID       int64
	ThreadID int64
	AuthorID int64
	Value    int
	Created  time.Time
	Updated  time.Time
}

type Threads []*Thread

func (threads Threads) GetAuthorIDs() (ids []int64) {
	for _, thread := range threads {
		ids = append(ids, thread.AuthorID)
	}
	return
}

func (threads Threads) GetAuthorIDsInterface() (ids []interface{}) {
	for _, thread := range threads {
		ids = append(ids, thread.AuthorID)
	}
	return
}
