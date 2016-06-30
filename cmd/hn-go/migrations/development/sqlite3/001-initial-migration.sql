-- +migrate Up
CREATE TABLE users(
       id integer primary key autoincrement,
       username varchar(255) not null,
       password varchar(255) not null,
       email varchar(255) not null,
       created timestamp default(datetime('now')),
       updated timestamp default(datetime('now'))
);

-- +migrate Down
DELETE TABLE users;
