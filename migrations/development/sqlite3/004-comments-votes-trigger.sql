-- +migrate Up

-- +migrate StatementBegin

-- When a new comment is created, a comment_votes record is inserted with a score of 1

CREATE TRIGGER comment_inserted AFTER INSERT ON comments
BEGIN
    INSERT INTO comment_votes (author_id, comment_id, score )  VALUES ( new.author_id, new.id, 1 );
END;
-- +migrate StatementEnd

