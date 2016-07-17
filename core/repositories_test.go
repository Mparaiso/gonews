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
	db := LoadFixtures(MigrateUp(GetDB(t), t), t)

	threadRepository := &gonews.ThreadRepository{DB: db, Logger: gonews.NewDefaultLogger(gonews.OFF)}
	threads, err := threadRepository.GetByAuthorID(1)
	Expect(t, err, nil)
	Expect(t, len(threads), 2, "len(threads)")
	Expect(t, threads[0].ID, int64(1), "threads[0].ID")
	Expect(t, threads[0].AuthorID, int64(1), "threads[0].AuthorID")
}

func TestThreadRepository_GetByIDWithComments(t *testing.T) {
	db := LoadFixtures(MigrateUp(GetDB(t), t), t)
	threadRepository := &gonews.ThreadRepository{DB: db, Logger: gonews.NewDefaultLogger(gonews.OFF)}
	thread, err := threadRepository.GetByIDWithComments(1)
	Expect(t, err, nil)
	Expect(t, thread.AuthorID, int64(1), "thread.AuthorID")
}

func TestCommentRepository_GetCommentsByAuthorID(t *testing.T) {
	db := LoadFixtures(MigrateUp(GetDB(t), t), t)
	count, authorId := 0, 1
	row := db.QueryRow("SELECT COUNT(ID) FROM comments_view WHERE AuthorID = ? ", int64(authorId))
	err := row.Scan(&count)
	Expect(t, err, nil)
	commentRepository := &gonews.CommentRepository{DB: db, Logger: gonews.NewDefaultLogger(gonews.OFF)}
	comments, err := commentRepository.GetCommentsByAuthorID(int64(authorId))
	Expect(t, err, nil)
	Expect(t, len(comments), count, "comments count")
}
