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

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/gorilla/sessions"
)

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
	// for each request, the container is then passed to all middlewares in the stack
	if appOptions.ContainerFactory == nil {
		appOptions.ContainerFactory = func() *Container {
			container := &Container{
				ContainerOptions: appOptions.ContainerOptions,
			}
			container.LoggerProvider = NewDefaultLoggerProvider(container.ContainerOptions.LoggerFactory,
				container.ContainerOptions.LogLevel,
				container.ContainerOptions.Debug)

			container.SessionProvider = NewDefaultSessionProvider(appOptions.ContainerOptions.Session.Name,
				container, container, container)

			container.CSRFGeneratorProvider = NewDefaultCSRFGeneratorProvider(container, container)

			container.TemplateProvider = NewDefaultTemplateProvider(container.ContainerOptions.TemplateDirectory,
				container.ContainerOptions.TemplateFileExtension,
				container.ContainerOptions.Debug, container)

			container.FormDecoderProvider = NewDefaultFormDecoderProvider(new(DefaultFormDecoder))

			return container
		}
	}
	// migrations
	if appOptions.Migrate == true {

	}
	// This is the default middleware stack each requests pass through all these middlewares
	// before being handled by a controller (which is also a middleware FYI )
	DefaultStack := &Stack{
		Middlewares: []Middleware{
			StopWatchMiddleware,   // Times how long it takes for the request to be handled
			LoggerMiddleware,      // Logs each request using the common log format
			SessionMiddleware,     // Initializes the session
			RefreshUserMiddleware, // Refresh an authenticated user if user.ID exists in session
			TemplateMiddleware,    // Configures template environment
		}, ContainerFactory: appOptions.ContainerFactory}

	Default := DefaultStack.Clone().Build()
	// Usef for authenticated routes
	AuthenticatedUsersOnly := DefaultStack.Clone().Push(AuthenticatedUserOnlyMiddleware).Build()
	app := http.NewServeMux()
	routes := Route{}

	// index
	app.HandleFunc(routes.StoriesByScore(), Default(NotFoundMiddleware, ThreadIndexController))

	app.HandleFunc(routes.NewComments(), Default(NewCommentsController))

	app.HandleFunc(routes.NewStories(), Default(NewestStoriesController))

	app.HandleFunc(routes.Thread(), Default(ThreadShowController))

	app.HandleFunc(routes.Reply(), AuthenticatedUsersOnly(CommentCreateController))

	app.HandleFunc(routes.ThreadsByDomain(), Default(ThreadByHostController))

	app.HandleFunc(routes.Login(), Default(LoginController))

	app.HandleFunc(routes.Logout(), Default(PostOnlyMiddleware, LogoutController))

	app.HandleFunc(routes.UserProfile(), Default(UserShowController))

	app.HandleFunc(routes.CreateStory(), AuthenticatedUsersOnly(SubmissionController))

	app.HandleFunc(routes.AuthorStories(), Default(ThreadListByAuthorIDController))

	app.HandleFunc(routes.AuthorComments(), Default(CommentsByAuthorController))

	app.HandleFunc(routes.Registration(), Default(PostOnlyMiddleware, RegistrationController))

	app.Handle(routes.Public(), http.StripPrefix(routes.Public(), http.FileServer(http.Dir(appOptions.PublicDirectory))))

	return app
}

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
			Debug:                 false,
			LogLevel:              INFO,
			Title:                 "gonews",
			Environment:           "development",
			Slogan:                "the news site for gophers",
			Description:           "gonews is a site where gophers publish and discuss news about the go language",
			DataSource:            "db.sqlite3",
			Driver:                "sqlite3",
			TemplateDirectory:     "templates",
			TemplateFileExtension: "tpl.html",
			Secret:                string(secret),
			CommentMaxDepth:       5,
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
	Migrate bool
	PublicDirectory,
	// Current App Environment : development,production,staging,testing
	Environment string
	ContainerOptions
	ContainerFactory func() *Container
}

// Route configures URIs
type Route struct{}

// StoriesByScore is the index
func (Route) StoriesByScore() string { return "/" }

// NewStories URI displays stories by age
func (Route) NewStories() string { return "/newest" }

// NewComments URI displays comments by age
func (Route) NewComments() string     { return "/newcomments" }
func (Route) Thread() string          { return "/item" }
func (Route) Reply() string           { return "/reply" }
func (Route) ThreadsByDomain() string { return "/from" }
func (Route) AuthorComments() string  { return "/threads" }
func (Route) AuthorStories() string   { return "/submitted" }
func (Route) Login() string           { return "/login" }
func (Route) Registration() string    { return "/register" }
func (Route) Public() string          { return "/public/" }
func (Route) Logout() string          { return "/logout" }
func (Route) UserProfile() string     { return "/user" }
func (Route) CreateStory() string     { return "/submit" }
