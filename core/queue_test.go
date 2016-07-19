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

package gonews_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/mparaiso/gonews/core"
)

// ExampleStack_first demonstrates the use of the middleware stack
func ExampleMiddlewareQueue_first() {
	context := ""
	stack := &gonews.MiddlewareQueue{Middlewares: []gonews.Middleware{
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
