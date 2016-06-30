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
       updated timestamp not null default(datetime('now'))
);

CREATE TABLE comments(
       id integer primary key autoincrement,
       parent_id integer references comments(id) ON DELETE CASCADE,
       thread_id integer references threads(id) ON DELETE CASCADE,
       content text not null,
       created timestamp not null default(datetime('now')),
       updated timestamp not null default(datetime('now'))
);

CREATE TABLE comment_votes(
       id integer primary key autoincrement,
       comment_id integer not null references comments(id),
       author_id integer not null references users(id),
       'value' integer ,
       created timestamp not null default(datetime('now')),
       updated timestamp not null default(datetime('now'))

);

CREATE TABLE thread_votes(
       id integer primary key autoincrement,
       thread_id integer not null references comments(id),
       author_id integer not null references users(id),
       'value' integer ,
       created timestamp not null default(datetime('now')),
       updated timestamp not null default(datetime('now'))
);

-- +migrate Down

DELETE TABLE roles;
DELETE TABLE users_roles;
DELETE INDEX users_roles_index;
DELETE TABLE threads;
DELETE TABLE comments;
DELETE TABLE comment_votes;
DELETE TABLE thread_votes;
