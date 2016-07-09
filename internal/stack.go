package gonews

import "net/http"

// Next is a function that call the next middleware or the handler
// if all middlewares have been called
type Next func()

// ContainerFactory creates a container
type ContainerFactory func() *Container

// Middleware is a middleware before an handler
type Middleware func(*Container, http.ResponseWriter, *http.Request, func())

// Handler handles an http request
type Handler Middleware

// Stack is a stack of handlers
type Stack struct {
	Middlewares []Middleware
	ContainerFactory
}

// Push add a new middleware to the stack
func (s *Stack) Push(m Middleware) *Stack {
	s.Middlewares = append(s.Middlewares, m)
	return s
}

// Shift prepends a new middleware to the stack
func (s *Stack) Shift(m Middleware) *Stack {
	s.Middlewares = append([]Middleware{m}, s.Middlewares...)
	return s
}

// Clone clones *Stack
func (s *Stack) Clone() *Stack {
	return &Stack{Middlewares: s.Middlewares[:], ContainerFactory: s.ContainerFactory}
}

// Build returns a function that returns a http.HandlerFunc
func (s *Stack) Build() func(...Handler) http.HandlerFunc {

	return func(handlers ...Handler) http.HandlerFunc {
		// copy all the middlewares
		middlewares := s.Middlewares[:]
		for _, handler := range handlers {
			middlewares = append(middlewares, Middleware(handler))
		}
		return func(rw http.ResponseWriter, r *http.Request) {
			if s.ContainerFactory == nil {
				s.ContainerFactory = func() *Container { return new(Container) }
			}
			container := s.ContainerFactory()
			rwn := &ResponseWriterExtra{ResponseWriter: rw, Request: r}
			var i int
			var next func()

			next = func() {
				if len(middlewares) > i {
					i++
					middlewares[i-1](container, rwn, r, next)

				} else if !rwn.IsResponseWritten() {
					// if next is called while handler has already been called
					// and no response has been written
					// then status No Content
					rwn.WriteHeader(204)
				}
			}
			next()

		}
	}
}
