-- +migrate Up

-- +migrate StatementBegin
CREATE VIEW IF NOT EXISTS threads_view AS 
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
	           )
	           t
	           LEFT JOIN
	           comments c ON c.thread_id = t.ID
	           GROUP BY t.id
-- +migrate StatementEnd		
		
			
-- +migrate Down
DROP VIEW IF EXISTS threads_view