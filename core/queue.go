//    Gonews is a webapp that provides a forum where users can post and discuss links
//
//    Copyright (C) 2016  mparaiso <mparaiso@online.fr>
//
//    This program is free software: you can redistribute it and/or modify
//    it under the terms of the GNU Affero General Public License as published
//    by the Free Software Foundation, either version 3 of the License, or
//    (at your option) any later version.
//
//    This program is distributed in the hope that it will be useful,
//    but WITHOUT ANY WARRANTY; without even the implied warranty of
//    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//    GNU Affero General Public License for more details.
//
//    You should have received a copy of the GNU Affero General Public License
//    along with this program.  If not, see <http://www.gnu.org/licenses/>.

package gonews

import "net/http"
import "runtime"
import "fmt"

// Next is a function that call the next middleware or the handler
// if all middlewares have been called
type Next func()

// ContainerFactory creates a container
type ContainerFactory func() *Container

// Middleware is a middleware. It can be transcient or final
type Middleware func(*Container, http.ResponseWriter, *http.Request, func())

// MiddlewareQueue is a stack of handlers
type MiddlewareQueue struct {
	Middlewares []Middleware
	ContainerFactory
}

// Push add a new middleware to the stack
func (s *MiddlewareQueue) Push(m Middleware) *MiddlewareQueue {
	s.Middlewares = append(s.Middlewares, m)
	return s
}

// Shift prepends a new middleware to the stack
func (s *MiddlewareQueue) Shift(m Middleware) *MiddlewareQueue {
	s.Middlewares = append([]Middleware{m}, s.Middlewares...)
	return s
}

// Clone clones *Stack
func (s *MiddlewareQueue) Clone() *MiddlewareQueue {
	var middlewares []Middleware
	for _, middleware := range s.Middlewares {
		middlewares = append(middlewares, middleware)
	}
	return &MiddlewareQueue{Middlewares: middlewares, ContainerFactory: s.ContainerFactory}
}

// Build returns a function that returns a http.HandlerFunc
func (s *MiddlewareQueue) Build() func(...Middleware) http.HandlerFunc {
	// copy all the middlewares
	var middlewares []Middleware
	for _, middleware := range s.Middlewares {
		middlewares = append(middlewares, middleware)
	}

	return func(handlers ...Middleware) http.HandlerFunc {

		var finalMiddlewareStack []Middleware
		for _, middleware := range middlewares {
			finalMiddlewareStack = append(finalMiddlewareStack, middleware)
		}

		for _, handler := range handlers {
			finalMiddlewareStack = append(finalMiddlewareStack, handler)
		}
		return func(rw http.ResponseWriter, r *http.Request) {
			if s.ContainerFactory == nil {
				s.ContainerFactory = func() *Container { return new(Container) }
			}
			container := s.ContainerFactory()
			rwe := &DefaultResponseWriterExtra{ResponseWriter: rw, Request: r}
			container.SetRequest(r)
			container.SetResponse(rwe)
			var i int
			var next func()
			defer func() {
				// handle potential panic
				err := recover()
				if err != nil {
					message := func() interface{} {
						if container.Debug() {
							return err
						}
						return http.StatusText(http.StatusInternalServerError)
					}()
					container.HTTPError(container.ResponseWriter(), container.Request(), 500, message)
					container.MustGetLogger().Error("recovered error \t", err)
					b := make([]byte, 1000)
					runtime.Stack(b, true)
					fmt.Printf("%s", b)
					return
				}
			}()
			next = func() {
				if len(finalMiddlewareStack) > i {
					i++
					finalMiddlewareStack[i-1](container, rwe, r, next)

				}
			}
			next()

		}
	}
}
