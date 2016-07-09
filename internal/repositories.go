// A repository persists models and queries the database for models
// This file centralize all repositories used in the application

package gonews

import (
	"database/sql"
	"fmt"
	"strings"
)

// UserRepository is a repository of users
type UserRepository struct {
	DB     *sql.DB
	Logger LoggerInterface
}

// Save persists a user
func (ur *UserRepository) Save(u *User) error {
	if u.ID == 0 {
		// user must be created
		command := "INSERT INTO users(username,email,password) VALUES(?,?,?);"
		ur.debug(command, u)
		result, err := ur.DB.Exec(command, u.Username, u.Email, u.Password)
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
func (ur *UserRepository) GetOneByEmail(email string) (user *User, err error) {
	query := `SELECT u.id,
  	u.username,
	u.password,
	u.email,
	u.created,
	u.updated 
	from users u
	WHERE u.email  = ? ;
  `
	ur.debug(query, email)
	row := ur.DB.QueryRow(query, email)
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
func (ur *UserRepository) GetOneByUsername(username string) (user *User, err error) {
	query := `SELECT u.id,
  	u.username,
	u.password,
	u.email,
	u.created,
	u.updated 
	from users u
	WHERE u.username  = ? ;
  `
	ur.debug(query, username)
	row := ur.DB.QueryRow(query, username)
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

// GetUserById returns a user, an error on error or nil if user not found
func (ur *UserRepository) GetById(id int64) (user *User, err error) {
	query := `SELECT 
	u.id AS ID,
	u.username AS Username,
	u.password AS Password,
	u.email AS Email,
	u.created AS Created,
	u.updated AS Updated
	FROM users u 
	WHERE u.id = ?`
	ur.debug(query, id)
	row := ur.DB.QueryRow(query, id)
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
	ur.debug(query, id)
	row = ur.DB.QueryRow(query, id)

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
	ur.debug(query, id)
	row = ur.DB.QueryRow(query, id)

	if err = row.Scan(&threadKarma); err != nil {
		return nil, err
	}
	user.Karma = threadKarma + commentKarma
	return
}

func (t UserRepository) debug(messages ...interface{}) {
	if t.Logger != nil {
		t.Logger.Debug(messages...)
	}
}

// RoleRepository is a repositorCreated y of roles
type RoleRepository struct{}

// ThreadRepository is a repository of threads
type ThreadRepository struct {
	DB     *sql.DB
	Logger LoggerInterface
}

func (t ThreadRepository) log(messages ...interface{}) {
	if t.Logger != nil {
		t.Logger.Debug(messages...)
	}
}

// GetByAuthorID returns threads filtered by AuthorID
func (t ThreadRepository) GetByAuthorID(id int64) (threads Threads, err error) {
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
	t.log(query, id)
	rows, err := t.DB.Query(query, id)
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
func (tr ThreadRepository) GetThreadByIDWithCommentsAndTheirAuthors(id int) (thread *Thread, err error) {
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
	tr.Logger.Debug(query, id)
	row := tr.DB.QueryRow(query, id)
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
	tr.Logger.Debug(query2, thread.AuthorID)
	row = tr.DB.QueryRow(query2, thread.AuthorID)
	author := new(User)
	err = MapRowToStruct([]string{"ID", "Username"}, row, author, true)
	if err != nil {
		return nil, err
	}
	// Comments
	thread.Author = author
	query3 := `SELECT comments.id AS ID,
	comments.content as Content, 
	comments.author_id AS AuthorID,
	u.username AS AuthorName,
	comments.created AS Created,
	comments.thread_id AS ThreadID,
	comments.parent_id AS ParentID,
	COUNT(comment_votes.score) as CommentScore  
	FROM comments 
	JOIN users u ON u.id = comments.author_id 
	LEFT JOIN comment_votes ON comment_votes.comment_id = comments.id
	WHERE comments.thread_id = ? 
	GROUP BY comments.id 
	ORDER BY CommentScore DESC, Created DESC ;`
	tr.Logger.Debug(query3, id)
	rows, err := tr.DB.Query(query3, id)
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
func (tr ThreadRepository) GetThreadsOrderedByVoteCount(limit, offset int) (threads Threads, err error) {
	query := `SELECT threads.id AS ID,threads.author_id AS AuthorID,threads.title AS Title,
	threads.created AS Created, threads.url as URL ,
	COUNT(thread_votes.id) AS Score FROM threads LEFT JOIN
	thread_votes ON thread_votes.thread_id = threads.id
	GROUP BY threads.id
	ORDER BY score DESC ,threads.created DESC
	LIMIT ? OFFSET ?;`
	defer tr.Logger.Debug(query, limit, offset)
	rows, err := tr.DB.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	err = MapRowsToSliceOfStruct(rows, &threads, true)
	if err != nil {
		return nil, err
	}

	ids := threads.GetAuthorIDsInterface()
	inClause := strings.TrimRight(strings.Repeat("?,", len(ids)), ",")
	queryCommentCount := fmt.Sprintf(`SELECT threads.id as ID,
	COUNT(comments.id) as CommentCount 
	FROM threads LEFT JOIN comments ON comments.thread_id = threads.id 
	WHERE threads.id IN(%s) 
	GROUP BY threads.id;`, inClause)
	defer tr.Logger.Debug(queryCommentCount, ids)
	rows, err = tr.DB.Query(queryCommentCount, ids...)
	type CommentCount struct {
		ID           int64
		CommentCount int
	}
	sliceOfcommentCount := []*CommentCount{}
	err = MapRowsToSliceOfStruct(rows, &sliceOfcommentCount, true)
	if err != nil {
		return nil, err
	}
	commentCountMap := map[int64]int{}
	for _, commentCount := range sliceOfcommentCount {
		commentCountMap[commentCount.ID] = commentCount.CommentCount
	}
	for _, thread := range threads {
		thread.CommentCount = commentCountMap[thread.ID]
	}
	query2 := fmt.Sprintf("SELECT username as Username, id as ID FROM users WHERE id IN(%s)", inClause)
	defer tr.Logger.Debug(query2, ids)

	rows, err = tr.DB.Query(query2, ids...)
	if err != nil {
		return nil, err
	}
	var users []*User
	err = MapRowsToSliceOfStruct(rows, &users, true)
	if err != nil {
		return nil, err
	}
	var userMap = map[int64]*User{}
	for _, user := range users {
		userMap[user.ID] = user
	}
	for _, thread := range threads {
		thread.Author = userMap[thread.AuthorID]
	}
	return
}

// CommentRepository is a repository of comments
type CommentRepository struct{}

// CommentVoteRepository is a repository of comment votes
type CommentVoteRepository struct{}

// ThreadVoteRepository is a repository of thread votes
type ThreadVoteRepository struct{}
