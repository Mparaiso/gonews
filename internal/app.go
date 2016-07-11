// @Copyright (c) 2016 mparaiso <mparaiso@online.fr>  All rights reserved.

package gonews

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path"

	// "github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

// DefaultContainerOptions returns the default ContainerOptions
// A closure is used to generate the function which allows us
// to have a few global variables ,like the session store or the db
var DefaultContainerOptions = func() func() ContainerOptions {
	//secret := securecookie.GenerateRandomKey(64)
	connection, connectionErr := sql.Open("sqlite3", "db.sqlite3")
	secret := []byte("some secret key for debugging purposes")
	sessionCookieStore := sessions.NewCookieStore(secret)
	sessionCookieStore.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   60 * 60 * 24,
		Domain:   "localhost",
	}

	return func() ContainerOptions {
		options := ContainerOptions{
			Debug:             false,
			LogLevel:          INFO,
			Title:             "gonews",
			Slogan:            "the news site for gophers",
			Description:       "gonews is a site where gophers publish and discuss news about the go language",
			DataSource:        "db.sqlite3",
			Driver:            "sqlite3",
			TemplateDirectory: "templates",
			Secret:            string(secret),
			Session: struct {
				Name         string
				StoreFactory func() (sessions.Store, error)
			}{
				Name: "go-news",
				StoreFactory: func() (sessions.Store, error) {
					return sessionCookieStore, nil
				},
			},
			ConnectionFactory: func() (*sql.DB, error) {
				return connection, connectionErr
			},
		}
		return options
	}
}()

// AppOptions gather all the configuration options
type AppOptions struct {
	PublicDirectory string
	ContainerOptions
	ContainerFactory func() *Container
}

// App is the application
type App struct {
	*http.ServeMux
}

// GetApp returns an application ready to be handled by a server
func GetApp(appOptions AppOptions) http.Handler {
	// Normalize appOptions
	if appOptions.PublicDirectory == "" {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		appOptions.PublicDirectory = path.Join(wd, "public")
	}
	// The containerFactory will be used to create a new container
	// for each request, the container is then passed to all middlewares
	if appOptions.ContainerFactory == nil {
		appOptions.ContainerFactory = func() *Container {
			container := &Container{
				ContainerOptions: appOptions.ContainerOptions,
			}
			container.SessionProvider = NewDefaultSessionProvider(appOptions.ContainerOptions.Session.Name, container, container, container)
			container.CSRFGeneratorProvider = NewDefaultCSRFGeneratorProvider(container, container)
			return container
		}
	}

	DefaultStack := &Stack{
		Middlewares: []Middleware{
			StopWatchMiddleware,   // Benchmarks the stack execution time
			LoggerMiddleware,      // Logs each request formatted by the common log format
			SessionMiddleware,     // Initialize the session
			RefreshUserMiddleware, // Refresh an authenticated user if user.ID exists in session
			TemplateMiddleware,    // Configures templates environment
		}, ContainerFactory: appOptions.ContainerFactory}
	// A middleware stack with request logging
	Default := DefaultStack.Clone().Build()
	AuthenticatedUsersOnly := DefaultStack.Clone().Push(AuthenticatedUserOnlyMiddleware).Build()
	// A middleware stack that extends Zero and handles requests for missing pages
	app := &App{http.NewServeMux()}
	// homepage
	app.HandleFunc("/", Default(NotFoundMiddleware, ThreadIndexController))
	// thread
	app.HandleFunc("/thread", Default(ThreadShowController))
	// login
	app.HandleFunc("/login", Default(LoginController))
	// logout
	app.HandleFunc("/logout", Default(PostOnlyMiddleware, LogoutController))
	// user
	app.HandleFunc("/user", Default(UserShowController))
	// submit
	app.HandleFunc("/submit", AuthenticatedUsersOnly(SubmissionController))
	// submitted
	app.HandleFunc("/submitted", Default(ThreadListByAuthorIDController))
	// registration
	app.HandleFunc("/register", Default(PostOnlyMiddleware, RegistrationController))
	//public static files
	app.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir(appOptions.PublicDirectory))))
	// not found
	return app
}
