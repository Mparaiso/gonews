// @Copyright (c) 2016 mparaiso <mparaiso@online.fr>  All rights reserved.

package gonews

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

// DefaultContainerOptions returns the default ContainerOptions
// A closure is used to generate the function which allows us
// to have a few global variables ,like the session store or the db
var DefaultContainerOptions = func() func() ContainerOptions {
	secret := securecookie.GenerateRandomKey(64)
	secretString := string(secret)
	sessionCookieStore := sessions.NewCookieStore(secret)
	connection, connectionErr := sql.Open("sqlite3", "db.sqlite3")
	return func() ContainerOptions {
		return ContainerOptions{
			Debug:             true,
			DataSource:        "db.sqlite3",
			Driver:            "sqlite3",
			TemplateDirectory: "templates",
			Secret:            secretString,
			SessionStoreFactory: func() (sessions.Store, error) {
				return sessionCookieStore, nil
			},
			ConnectionFactory: func() (*sql.DB, error) {
				return connection, connectionErr
			},
		}
	}
}()

// AppOptions gather all the configuration options
type AppOptions struct {
	PublicDirectory string
}

// App is the application
type App struct {
	*http.ServeMux
}

// GetApp returns an application ready to be handled by a server
func GetApp(options ContainerOptions, appOptions AppOptions) http.Handler {
	// Normalize appOptions
	if appOptions.PublicDirectory == "" {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		appOptions.PublicDirectory = path.Join(wd, "public")
	}
	containerFactory := func() *Container {
		return &Container{ContainerOptions: options}
	}
	DefaultStack := &Stack{
		Middlewares: []Middleware{
			StopWatchMiddleware,
			LoggingMiddleware,
			SessionMiddleware,
			CSRFMiddleWare,
			TemplateMiddleware,
		}, ContainerFactory: containerFactory}
	// A middleware stack with request logging
	Default := DefaultStack.Build()
	// A middleware stack that extends Zero and handles requests for missing pages
	Home := DefaultStack.Clone().Push(NotFoundMiddleware).Build()
	app := &App{http.NewServeMux()}
	// homepage
	app.HandleFunc("/", Home(ThreadIndexController))
	// thread
	app.HandleFunc("/thread", Default(ThreadShowController))
	// login
	app.HandleFunc("/login", Default(LoginController))
	// user
	app.HandleFunc("/user", Default(UserShowController))
	// submitted user stories
	app.HandleFunc("/submitted", Default(ThreadListByAuthorIDController))
	// registration
	app.HandleFunc("/register", Default(PostOnlyMiddleware, RegistrationController))
	//public static files
	app.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir(appOptions.PublicDirectory))))
	// not found
	return app
}
