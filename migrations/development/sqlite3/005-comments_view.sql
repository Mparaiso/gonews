-- +migrate Up

-- +migrate StatementBegin

-- comments_view 

CREATE VIEW IF NOT EXISTS comments_view AS 
	SELECT c.id AS ID,
	       c.content AS Content,
	       c.author_id AS AuthorID,
	       u.username AS AuthorName,
	       c.created AS Created,
		   c.updated AS Updated,
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
	 GROUP BY c.id ;
	
-- +migrate StatementEnd		
		
			
-- +migrate Down
DROP VIEW IF EXISTS comments_view