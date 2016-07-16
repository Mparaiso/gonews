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

// A repository persists models and queries the database for models
// This file centralize all repositories used in the application

package gonews

import (
	"database/sql"
	"fmt"
)

// Query is an SQL Query
type Query string

// UserRepository is a repository of users
type UserRepository struct {
	DB     *sql.DB
	Logger LoggerInterface
}

// Save persists a user
func (repository *UserRepository) Save(u *User) error {
	if u.ID == 0 {
		// user must be created
		command := "INSERT INTO users(username,email,password) VALUES(?,?,?);"
		repository.debug(command, u)
		result, err := repository.DB.Exec(command, u.Username, u.Email, u.Password)
		if err != nil {
			return err
		}
		u.ID, err = result.LastInsertId()
		if err != nil {
			return err
		}
		return nil
	}
	// user must be updated

	return fmt.Errorf("user update Not implemented ")
}

// GetOneByEmail gets one user by his email
func (repository *UserRepository) GetOneByEmail(email string) (user *User, err error) {
	query := `SELECT u.id,
  	u.username,
	u.password,
	u.email,
	u.created,
	u.updated 
	from users u
	WHERE u.email  = ? ;
  `
	repository.debug(query, email)
	row := repository.DB.QueryRow(query, email)
	user = new(User)
	err = MapRowToStruct([]string{"ID", "Password", "Email", "Created", "Updated"}, row, user, true)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, nil
		default:
			return nil, err
		}
	}
	return
}

// GetOneByUsername gets one user by his name
func (repository *UserRepository) GetOneByUsername(username string) (user *User, err error) {
	query := `SELECT u.id,
  	u.username,
	u.password,
	u.email,
	u.created,
	u.updated 
	from users u
	WHERE u.username  = ? ;
  `
	repository.debug(query, username)
	row := repository.DB.QueryRow(query, username)
	user = new(User)
	err = MapRowToStruct([]string{"ID", "Username", "Password", "Email", "Created", "Updated"}, row, user, true)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, nil
		default:
			return nil, err
		}
	}
	return
}

// GetByID returns a user, an error on error or nil if user not found
func (repository *UserRepository) GetByID(id int64) (user *User, err error) {
	query := `SELECT 
	u.id AS ID,
	u.username AS Username,
	u.password AS Password,
	u.email AS Email,
	u.created AS Created,
	u.updated AS Updated
	FROM users u 
	WHERE u.id = ?`
	repository.debug(query, id)
	row := repository.DB.QueryRow(query, id)
	user = new(User)
	err = MapRowToStruct([]string{"ID", "Username", "Password", "Email", "Created", "Updated"}, row, user, true)
	if err != nil {
		return
	}
	if err == sql.ErrNoRows {
		return nil, err
	}
	var threadKarma, commentKarma int
	query = `SELECT coalesce(SUM(comment_votes.score),0) as KarmaComments
  	FROM comment_votes
    JOIN
    comments ON comments.id = comment_votes.comment_id
    JOIN
    users ON users.id = comments.author_id
 	WHERE users.id = ?;
	`
	repository.debug(query, id)
	row = repository.DB.QueryRow(query, id)

	if err = row.Scan(&commentKarma); err != nil {
		return nil, err
	}
	query = `SELECT coalesce(SUM(thread_votes.score),0) as Karmathreads
  	FROM thread_votes
    JOIN
    threads ON threads.id = thread_votes.thread_id
    JOIN
    users ON users.id = threads.author_id
 	WHERE users.id = ?;
	`
	repository.debug(query, id)
	row = repository.DB.QueryRow(query, id)

	if err = row.Scan(&threadKarma); err != nil {
		return nil, err
	}
	user.Karma = threadKarma + commentKarma
	return
}

func (repository UserRepository) debug(messages ...interface{}) {
	if repository.Logger != nil {
		repository.Logger.Debug(messages...)
	}
}

// RoleRepository is a repositorCreated y of roles
type RoleRepository struct{}

// ThreadRepository is a repository of threads
type ThreadRepository struct {
	DB     *sql.DB
	Logger LoggerInterface
}

func (repository ThreadRepository) log(messages ...interface{}) {
	if repository.Logger != nil {
		repository.Logger.Debug(messages...)
	}
}

// Create creates  an thread in the database
func (repository ThreadRepository) Create(thread *Thread) error {
	command := "INSERT INTO threads(title,url,content,author_id) values(?,?,?,?);"
	repository.Logger.Debug(command, thread)
	result, err := repository.DB.Exec(command, thread.Title, thread.URL, thread.Content, thread.AuthorID)

	if err == nil {
		thread.ID, err = result.LastInsertId()
		// The following part is handled automatically by the DB with a TRIGGER
		// so it is commented for now
		// if err == nil {

		// 	threadVoteRepository := &ThreadVoteRepository{t.DB, t.Logger}
		// 	_, err = threadVoteRepository.Create(&ThreadVote{AuthorID: thread.AuthorID, ThreadID: thread.ID, Score: 1})
		// }
	}

	return err
}

// GetWhereURLLike returns threads where url like pattern
func (repository ThreadRepository) GetWhereURLLike(pattern string) (threads Threads, err error) {
	query := `SELECT t.id as ID ,
       t.title AS Title,
       t.url AS URL,
       t.created AS Created,
       t.updated AS Updated,
       t.author_id AS AuthorID,
	   t.username AS AuthorName,
       COALESCE(SUM(thread_votes.score),0) AS Score,
       CommentCount
  	FROM (
           SELECT threads.id ,
                  threads.title,
                  threads.url,
                  threads.created,
                  threads.updated,
                  threads.author_id,
				  users.username,
                  COALESCE(COUNT(comments.id), 0) AS CommentCount
             FROM threads,users
                  LEFT JOIN
                  comments ON comments.thread_id = threads.id
            WHERE threads.url LIKE ? 
				  AND users.id = threads.author_id
            GROUP BY threads.id
            ORDER BY threads.created DESC
    ) t
    LEFT JOIN
    thread_votes ON thread_votes.thread_id = t.id
 	GROUP BY t.id;`
	repository.Logger.Debug(query, pattern)
	var rows *sql.Rows
	rows, err = repository.DB.Query(query, pattern)
	if err == nil {
		err = MapRowsToSliceOfStruct(rows, &threads, true)
		if err == nil {
			for _, thread := range threads {
				thread.Author = new(User)
				thread.Author.Username = thread.AuthorName
				thread.Author.ID = thread.AuthorID
			}
		}
	}
	return
}

// GetByAuthorID returns threads filtered by AuthorID
func (repository ThreadRepository) GetByAuthorID(id int64) (threads Threads, err error) {
	// we query the database, first by search threads by author_id with the commentcount
	// then by aggregating the sum of thread_votes.score
	// TODO refactor as a view in the database
	query := `SELECT t.id as ID ,
       t.title AS Title,
       t.url AS URL,
       t.created AS Created,
       t.updated AS Updated,
       t.author_id AS AuthorID,
       COALESCE(SUM(thread_votes.score),0) AS Score,
       CommentCount
  	FROM (
           SELECT threads.id ,
                  threads.title,
                  threads.url,
                  threads.created,
                  threads.updated,
                  threads.author_id,
                  COALESCE(COUNT(comments.id), 0) AS CommentCount
             FROM threads
                  LEFT JOIN
                  comments ON comments.thread_id = threads.id
            WHERE threads.author_id = 1
            GROUP BY threads.id
            ORDER BY threads.created DESC
    ) t
    LEFT JOIN
    thread_votes ON thread_votes.thread_id = t.id
 	GROUP BY t.id;`
	repository.log(query, id)
	rows, err := repository.DB.Query(query, id)
	if err != nil {
		return nil, err
	}
	err = MapRowsToSliceOfStruct(rows, &threads, true)
	if err != nil {
		return nil, err
	}
	return
}

// GetThreadByIDWithCommentsAndTheirAuthors gets a threas with its comments
func (repository ThreadRepository) GetThreadByIDWithCommentsAndTheirAuthors(id int) (thread *Thread, err error) {
	// Thread
	query := `
	SELECT threads.id AS ID,
	threads.title AS Title,threads.url AS URL, threads.created AS Created , 
	COUNT(comments.id) AS CommentCount,
	threads.author_id AS AuthorID 
	FROM threads 
	LEFT JOIN comments ON comments.thread_id = threads.id
	WHERE threads.id = ? 
	GROUP BY threads.id;`
	repository.Logger.Debug(query, id)
	row := repository.DB.QueryRow(query, id)
	thread = new(Thread)
	err = MapRowToStruct([]string{"ID", "Title", "URL", "Created",
		"CommentCount", "AuthorID"}, row, thread, true)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	// Author
	query2 := `SELECT users.id AS ID,users.username AS Username
	FROM users WHERE users.id = ? ;
	`
	repository.Logger.Debug(query2, thread.AuthorID)
	row = repository.DB.QueryRow(query2, thread.AuthorID)
	author := new(User)
	err = MapRowToStruct([]string{"ID", "Username"}, row, author, true)
	if err != nil {
		return nil, err
	}
	// Comments
	thread.Author = author
	query3 := `
		SELECT c.id AS ID,
			c.content AS Content,
			c.author_id AS AuthorID,
			u.username AS AuthorName,
			c.created AS Created,
			c.thread_id AS ThreadID,
			c.parent_id AS ParentID,
			COUNT(cv.score) AS CommentScore,
			t.Title AS ThreadTitle
		FROM comments c
			JOIN
			users u ON u.id = c.author_id
			JOIN
			threads t ON t.id = c.thread_id
			LEFT JOIN
			comment_votes cv ON cv.comment_id = c.id
		WHERE c.thread_id = ?
		GROUP BY c.id
		ORDER BY CommentScore DESC,
				Created DESC;
				`
	repository.Logger.Debug(query3, id)
	rows, err := repository.DB.Query(query3, id)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	err = MapRowsToSliceOfStruct(rows, &thread.Comments, true)
	if err != nil {
		return nil, err
	}

	return
}

// GetThreadsOrderedByVoteCount returns threads ordered by thread vote count
func (repository ThreadRepository) GetSortedByScore(limit, offset int) (threads Threads, err error) {
	query := `
		SELECT t.ID,
		       t.AuthorID,
		       t.Title,
		       t.Created,
		       t.URL,
		       t.Score,
		       t.AuthorName,
		       coalesce(COUNT(c.id), 0) AS CommentCount
		FROM (
			SELECT threads.id AS ID,
			       threads.author_id AS AuthorID,
			       threads.title AS Title,
			       threads.created AS Created,
			       threads.url AS URL,
			       u.username AS AuthorName,
			       coalesce(SUM(thread_votes.score), 0) AS Score
			FROM threads
			JOIN
			    users u ON u.id = threads.author_id
			LEFT JOIN
				thread_votes ON thread_votes.thread_id = threads.id
			 GROUP BY threads.id
       	) t
       	LEFT JOIN
       		comments c ON c.thread_id = t.ID
       	GROUP BY t.ID
 		ORDER BY t.Score DESC, t.Created DESC
		LIMIT ? OFFSET ? ;`
	var (
		rows *sql.Rows
	)
	repository.Logger.Debug(query, limit, offset)
	rows, err = repository.DB.Query(query, limit, offset)
	if err == nil {
		err = MapRowsToSliceOfStruct(rows, &threads, true)
	}
	return
}

// GetOrderedByCreatedDesc returns threads ordered by age DESC
func (repository ThreadRepository) GetNewest() (threads Threads, err error) {
	query := `
		SELECT t.id AS ID,
			t.title AS Title,
			t.url AS URL,
			t.created AS Created,
			t.score AS Score,
			t.author_id AS AuthorID,
			t.username AS AuthorName,
			coalesce(COUNT(c.id), 0) AS CommentCount
		FROM (
				SELECT t.id,
						t.title,
						t.url,
						t.created,
						t.author_id,
						u.username,
						coalesce(SUM(tv.score), 0) AS score
					FROM threads t,
						users u
						LEFT JOIN
						thread_votes tv ON tv.thread_id = t.id
						WHERE t.author_id = u.id
					GROUP BY t.id
			) t
			LEFT JOIN
			comments c ON c.thread_id = t.id
		GROUP BY t.id
		ORDER BY t.created DESC ,t.score DESC;
		`
	repository.Logger.Debug(query)
	var (
		rows *sql.Rows
	)
	rows, err = repository.DB.Query(query)
	if err == nil {
		err = MapRowsToSliceOfStruct(rows, &threads, true)
		if err == nil || err == sql.ErrNoRows {
			err = nil
			return
		}
	}
	return
}

// CommentRepository is a repository of comments
type CommentRepository struct {
	*sql.DB
	Logger LoggerInterface
}

// GetNewestComments returns comments sorted by date of creation
func (repository *CommentRepository) GetNewestComments() (comments Comments, err error) {
	query := `SELECT 
			c.id AS ID,
			c.parent_id AS ParentID,
			c.thread_id AS ThreadID,
			c.author_id AS AuthorID,
			c.content AS Content,
			c.created AS Created,
			c.updated AS Updated,
			coalesce(SUM(comment_votes.score),0) AS Score,
			u.username AS AuthorName,
			t.Title AS ThreatTitle
		FROM comments c 
		JOIN threads t ON t.id = c.thread_id
		JOIN users u ON u.id = c.author_id
		LEFT JOIN comment_votes ON comment_votes.comment_id = c.id
		WHERE c.author_id = u.id
		GROUP BY c.id
		ORDER BY c.created DESC;`
	repository.Logger.Debug(query)
	var (
		rows *sql.Rows
	)
	rows, err = repository.DB.Query(query)
	if err == nil {
		err = MapRowsToSliceOfStruct(rows, &comments, true)
		if err == nil || err == sql.ErrNoRows {
			err = nil
			return
		}
	}
	return
}

// GetByID gets a comment by ID
func (repository *CommentRepository) GetByID(id int64) (comment *Comment, err error) {
	query := `
	SELECT  
		comments.id AS ID,
		comments.parent_id AS ParentID,
		comments.thread_id AS ThreadID,
		comments.author_id AS AuthorID,
		comments.content AS Content,
		comments.created AS Created,
		comments.updated AS Updated,
		coalesce(SUM(comment_votes.score),0) AS Score,
		users.username AS AuthorName 
	FROM 
		comments, users
	LEFT JOIN 
		comment_votes ON comment_votes.comment_id = comments.id
	WHERE 
		comments.id == ? 
		AND comments.author_id = users.id
	GROUP BY comments.id 
	LIMIT 1 ;`
	repository.Logger.Debug(query, id)
	row := repository.DB.QueryRow(query, id)
	comment = new(Comment)
	err = MapRowToStruct([]string{"ID", "ParentID", "ThreadID", "AuthorID", "Content", "Created", "Updated", "Score", "AuthorName"}, row, comment, true)
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	default:
		return comment, err
	}
}

// Create creates an new comment
func (repository *CommentRepository) Create(comment *Comment) error {
	command := `INSERT INTO comments(parent_id,thread_id,author_id,content)
		VALUES(?,?,?,?);`
	repository.Logger.Debug(command, comment)
	result, err := repository.DB.Exec(command,
		comment.ParentID, comment.ThreadID, comment.AuthorID, comment.Content,
	)
	if err == nil {
		if comment.ID, err = result.LastInsertId(); err == nil {
			return nil
		}
	}
	return err
}

// GetCommentsByAuthorID returns comments by author_id
func (repository *CommentRepository) GetCommentsByAuthorID(id int64) (comments Comments, err error) {
	var (
		rows *sql.Rows
	)
	query := `SELECT c.id AS ID,
				c.parent_id AS ParentID,
				c.author_id AS AuthorID,
				u.username AS AuthorName,
				c.thread_id AS ThreadID,
				c.content AS Content,
				c.created AS Created,
				c.updated AS Updated,
				coalesce(SUM(cv.score),0) AS CommentScore
			FROM comments c
				JOIN users u ON  c.author_id = u.id
				LEFT JOIN
				comment_votes cv ON cv.comment_id = c.id
			WHERE c.author_id = ? 
				
			GROUP BY c.id
			ORDER BY c.created DESC;`
	repository.Logger.Debug(query, id)
	rows, err = repository.DB.Query(query, id)
	if err == nil {
		err = MapRowsToSliceOfStruct(rows, &comments, true)
		if err == nil {
			return
		}
	}
	return
}

// CommentVoteRepository is a repository of comment votes
type CommentVoteRepository struct{}

// ThreadVoteRepository is a repository of thread votes
type ThreadVoteRepository struct {
	DB     *sql.DB
	Logger LoggerInterface
}

// Create creates a new thread vote
func (repository *ThreadVoteRepository) Create(threadVote *ThreadVote) (i int64, err error) {
	query := "INSERT INTO thread_votes(thread_id,author_id,score) values(?,?,?)"
	repository.Logger.Debug(query, threadVote)
	result, err := repository.DB.Exec(query, threadVote.ThreadID, threadVote.AuthorID, threadVote.Score)
	if err == nil {
		if i, err = result.LastInsertId(); err == nil {
			return i, nil
		}
	}
	return 0, err
}
