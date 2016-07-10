package gonews_test

import (
	"testing"

	"github.com/mparaiso/go-news/internal"
)

func Test_Comments_GetTree(t *testing.T) {
	comments := &gonews.Comments{{ID: 1, ParentID: 0}, {ID: 2, ParentID: 1}, {ID: 3, ParentID: 1}}
	commentTree := comments.GetTree()
	rootComment := commentTree[0]
	if rootComment == nil {
		t.Fatal("rootComment should not be nil")
	}
	if expected, actual := 2, len(rootComment.Children); expected != actual {
		t.Fatalf("%s : expect %v, got %v .", "root comment count", expected, actual)
	}
}

func TestUser_CreateSecurePassword(t *testing.T) {
	// Set up
	user := &gonews.User{}
	password := "thepassword"
	// Test
	err := user.CreateSecurePassword(password)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUser_Authenticate(t *testing.T) {
	// Set up
	user := &gonews.User{}
	password := "the password"
	// Test
	user.CreateSecurePassword(password)

	if err := user.Authenticate(password); err != nil {
		t.Fatal(err)
	}
}
