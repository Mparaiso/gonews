package main_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/mparaiso/go-news/internal"
)

// ExampleStack_first demonstrates the use of the middleware stack
func ExampleStack_first() {
	context := ""
	stack := &gonews.Stack{Middlewares: []gonews.Middleware{
		func(c *gonews.Container, rw http.ResponseWriter, r *http.Request, next func()) {
			context += "first middleware-"
			next()
			context += "end"
		},
		func(c *gonews.Container, rw http.ResponseWriter, r *http.Request, next func()) {
			context += "second middleware-"
			next()
		},
	}}
	middleware := stack.Build()
	server := http.NewServeMux()
	server.HandleFunc("/", middleware(func(c *gonews.Container, rw http.ResponseWriter, r *http.Request, next func()) {
		context += "main handler-"
		rw.Write([]byte("done"))
	}))
	testServer := httptest.NewServer(server)
	testServer.Config.WriteTimeout = 1000 * time.Millisecond
	defer testServer.Close()
	url := testServer.URL + "/"

	_, err := http.Get(url)
	fmt.Println(err)
	fmt.Println(context)

	// Output:
	// <nil>
	// first middleware-second middleware-main handler-end

}
