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
	"time"

	"net/url"

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
	// db columns
	ID       int64
	Title    string
	URL      string
	Content  string
	Created  time.Time
	Updated  time.Time
	AuthorID int64

	// Author is the author of the thread
	Author *User

	// Comments in the thread
	Comments Comments

	// virtual fields
	AuthorName   string
	CommentCount int
	Score        int
}

// GetURLHost returns the host of the thread url
func (t Thread) GetURLHost() (string, error) {

	u, err := url.Parse(t.URL)
	if err == nil {
		return u.Host, err
	}
	return "", err
}

// Comment is a comment in a thread
type Comment struct {
	ID       int64
	ParentID int64
	AuthorID int64

	ThreadID     int64
	Content      string
	CommentScore int
	Created      time.Time
	Updated      time.Time

	// virtual fields
	AuthorName  string
	Depth       int
	Children    Comments
	ThreadTitle string
}

// HasChildren return true is the comment has child comments
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
				subComment.Depth += 1
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
	Score    int64
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
