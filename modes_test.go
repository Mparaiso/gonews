package hn_test

import "testing"
import "github.com/mparaiso/hn-go"

func TestUser_CreateSecurePassword(t *testing.T) {
	// Set up
	user := &hn.User{}
	password := "thepassword"
	// Test
	err := user.CreateSecurePassword(password)
	if err != nil {
		t.Fatal(err)
	}


func TestUser_Authenticate(t *testing.T) {
	// Set up
	user := &hn.User{}
	password := "the password"
	// Test
	user.CreateSecurePassword(password)

	if err := user.Authenticate(password); err != nil {
		t.Fatal(err)
	}
}
