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

package gonews_test

import (
	"testing"

	"github.com/mparaiso/gonews/core"
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
