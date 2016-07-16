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

import "testing"
import gonews "github.com/mparaiso/gonews/core"

func TestThreadRepository_GetByAuthorID(t *testing.T) {
	db := MigrateUp(GetDB(t), t)
	threadRepository := &gonews.ThreadRepository{DB: db, Logger: gonews.NewDefaultLogger(gonews.OFF)}
	threads, err := threadRepository.GetByAuthorID(1)
	if err != nil {
		t.Fatal(err)
	}
	if expected, got := 2, len(threads); expected != got {
		t.Fatalf("threads length: expected '%v' , got '%v'", expected, got)
	}
	// t.Logf("%#v %#v", threads[0], threads[1])
	if expected, got := int64(1), threads[0].ID; expected != got {
		t.Fatalf("threads[0].ID : expected '%v' , got '%v' ", expected, got)
	}
	if expected, got := int64(1), threads[0].AuthorID; expected != got {
		t.Fatalf("threads[0].AuthorID: expected '%v' , got '%v'", expected, got)
	}
}
