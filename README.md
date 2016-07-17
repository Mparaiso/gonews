Gonews
======

[![Build Status](https://travis-ci.org/Mparaiso/gonews.svg?branch=master)](https://travis-ci.org/Mparaiso/gonews) 

[![Go Report Card](https://goreportcard.com/badge/github.com/Mparaiso/gonews)](https://goreportcard.com/report/github.com/Mparaiso/gonews)

author: mparaiso <mparaiso@online.fr>

#### Demonstration 

[gonews.herokuapp.com](https://gonews.herokuapp.com)

license: GNU AGPLv3

Gonews, a hacker news clone written in Go

This is a work in progress

Gonews is a hacker news like forum application where users can share links
and comment on these links 

[news.ycombinator.com](Hacker news)

Roadmap

- [ ] Documentation
- [x] Newest stories
- [x] Stories pagination
- [ ] Comments pagination
- [x] Most upvoted stories
- [x] New Comments
- [x] Creating Accounts
- [x] User profile
- [x] Signing in
- [x] Signing out
- [x] Replying to Comments
- [ ] Updating comments
- [ ] Upvoting comments
- [x] Submitting Stories
- [ ] Upvoting stories
- [ ] Administration
- [x] YAML configuration
- [x] sqlite support
- [ ] mysql support
- [ ] postgresql support

#### Getting Started

##### Installation

requirements: 
	
	-	go 1.6
	- 	sqlite3

in the command line :
	
	go get github.com/mparaiso/gonews
	
navigate to $GOPATH/src/github.com/mparaiso/gonews directory

	go build
	
Launch the server with the following commands:

	./gonews start -migrate -loadfixtures -port=8080
	
It will create the database, load some sample data and start 
the server on port 8080

To get some help on available options :

	./gonews help
	
	


