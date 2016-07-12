-- +migrate Up
CREATE TABLE roles (
       id integer primary key autoincrement,
       name varchar(255) not null
);

CREATE TABLE users_roles(
       id integer primary key autoincrement,
       user_id integer not null REFERENCES users(id) ON DELETE CASCADE,
       role_id integer not null REFERENCES  roles(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX users_roles_index ON users_roles(user_id,role_id);

CREATE TABLE threads(
       id integer primary key autoincrement,
       title varchar(255) not null,
       url varchar(255) not null,
       created timestamp not null default(datetime('now')),
       updated timestamp not null default(datetime('now')),
	   author_id int REFERENCES users(id) ON DELETE CASCADE,
	   content text 
);

-- comments

CREATE TABLE comments(
       id integer primary key autoincrement,
       parent_id integer default(0) references comments(id) ON DELETE CASCADE,
       thread_id integer default(0) references threads(id) ON DELETE CASCADE,
	   author_id integer default(0) references users(id),
       content text not null,
       created timestamp not null default(datetime('now')),
       updated timestamp not null default(datetime('now'))
);

CREATE TABLE comment_votes(
       id integer primary key autoincrement,
       comment_id integer not null references comments(id),
       author_id integer not null references users(id),
       score integer not null default(0),
       created timestamp not null default(datetime('now')),
       updated timestamp not null default(datetime('now'))

);

CREATE TABLE thread_votes(
       id integer primary key autoincrement,
       thread_id integer not null references threads(id),
       author_id integer not null references users(id),
       score integer not null default(0),
       created timestamp not null default(datetime('now')),
       updated timestamp not null default(datetime('now'))
);

CREATE UNIQUE INDEX thread_votes_index ON thread_votes(thread_id,author_id);

-- +migrate StatementBegin

-- When a new threads record is created, a new thread_votes record is automatically added with the thread.author_id and the thread.id
-- @see https://www.sqlite.org/lang_createtrigger.html

CREATE TRIGGER thread_inserted AFTER INSERT ON threads
BEGIN
    INSERT INTO thread_votes (author_id, thread_id, score )  VALUES ( new.author_id, new.id, 1 );
END;
-- +migrate StatementEnd


-- +migrate Down

DROP TABLE roles;
DROP TABLE users_roles;
DROP INDEX IF EXISTS users_roles_index;
DROP TABLE threads;
DROP TABLE comments;
DROP TABLE comment_votes;
DROP TABLE thread_votes;
