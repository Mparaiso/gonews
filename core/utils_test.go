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

package gonews_test

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"path/filepath"
	"time"

	"testing"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mparaiso/gonews/core"
	"github.com/rubenv/sql-migrate"

	"flag"
	"net/url"
	"runtime"
	"strings"
)

//
// HELPERS AND FIXTURES
//

// DEBUG will allows the test to output additional informations
var DEBUG = false

var DRIVER = "sqlite3"

var FORM_MIME_TYPE = "application/x-www-form-urlencoded"

// Directory is the current directory
var Directory = func() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return dir
}()

// Allows arguments to be passed to test
// ex: go test -args -debug
func TestMain(m *testing.M) {
	debug := flag.Bool("debug", DEBUG, "debug the test suite")
	driver := flag.String("driver", DRIVER, "database driver")
	flag.Parse()
	DEBUG = *debug
	DRIVER = *driver
	os.Exit(m.Run())
}

// GetDB gets the db connection
func GetDB(t *testing.T) *sql.DB {
	db, err := sql.Open(DRIVER, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	return db
}

// MigrateUp executes db migrations
func MigrateUp(db *sql.DB, t *testing.T) *sql.DB {
	_, err := migrate.Exec(db, DRIVER, migrate.FileMigrationSource{"./../migrations/development/" + DRIVER}, migrate.Up)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

// LoadFixtures loads test fixtures
func LoadFixtures(db *sql.DB, t *testing.T) *sql.DB {
	migrationFile, err := os.Open("./../testdata/fixtures/" + DRIVER + "/fixtures.sql")
	if err != nil {
		t.Fatal(err)
	}
	buffer := new(bytes.Buffer)
	_, err = buffer.ReadFrom(migrationFile)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(buffer.String())
	if err != nil {
		t.Fatal(err)
	}
	return db
}

// GetContainerOptions returns container options for tests
func GetContainerOptions(db *sql.DB) gonews.ContainerOptions {
	options := gonews.DefaultContainerOptions()
	options.Debug = DEBUG
	options.TemplateDirectory = "./../" + options.TemplateDirectory
	options.ConnectionFactory = func() (*sql.DB, error) {
		return db, nil
	}
	options.LogLevel = gonews.OFF
	return options
}

// GetServer sets up the test server with an optional db and returns the test server
func GetServer(t *testing.T, dbs ...*sql.DB) *httptest.Server {
	// Set Up
	var db *sql.DB
	if len(dbs) == 0 {
		db = GetDB(t)
	} else {
		db = dbs[0]
	}
	MigrateUp(db, t)
	LoadFixtures(db, t)
	app := gonews.GetApp(gonews.AppOptions{ContainerOptions: GetContainerOptions(db)})
	server := httptest.NewServer(app)

	logger := &log.Logger{}
	logger.SetOutput(os.Stdout)
	server.Config.ErrorLog = logger
	server.Config.WriteTimeout = 3 * time.Second
	http.DefaultClient.Jar = nil
	return server
}

// LoginUserHelper logs a user before executing a test
func LoginUser(t *testing.T) (*sql.DB, *httptest.Server, *gonews.User, error) {
	// GetServer
	db := GetDB(t)
	server := GetServer(t, db)
	unencryptedPassword := "password"
	user := &gonews.User{Username: "mike_doe", Email: "mike_doe@acme.com"}
	user.CreateSecurePassword(unencryptedPassword)
	result, err := db.Exec("INSERT INTO users(username,email,password) values(?,?,?);", user.Username, user.Email, user.Password)
	if err != nil {
		t.Fatal(err)
	}

	if n, err := result.RowsAffected(); err != nil || n != 1 {
		t.Fatal(n, err)
	}
	if l, err := result.LastInsertId(); err != nil {
		t.Fatal(err)
	} else {
		user.ID = l
	}

	// @see https://golang.org/pkg/net/http/cookiejar/
	// @see http://stackoverflow.com/questions/18414212/golang-how-to-follow-location-with-cookie
	http.DefaultClient.Jar, err = cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}

	res, err := http.Get(server.URL + "/login")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		t.Fatal(err)
	}
	selection := doc.Find("input[name='login_csrf']")

	csrf, ok := selection.First().Attr("value")
	if !ok {
		t.Fatal("csrf not found in HTML document", selection, ok)
	}
	if strings.Trim(csrf, " ") == "" {
		t.Fatal("csrf not found")
	}
	formValues := url.Values{
		"login_username": {user.Username},
		"login_password": {unencryptedPassword},
		"login_csrf":     {csrf},
	}
	res, err = http.Post(server.URL+"/login", "application/x-www-form-urlencoded", strings.NewReader(formValues.Encode()))
	defer res.Body.Close()

	if err != nil {
		t.Fatal(err)
	}

	if expected, got := 200, res.StatusCode; expected != got {
		t.Fatalf("POST /login status : expected '%v' got '%v'", expected, got)
	}

	doc, err = goquery.NewDocumentFromResponse(res)

	if err != nil {
		t.Fatal(err)
	}
	selection = doc.Find(".current-user")
	if expected, got := 1, selection.Length(); expected != got {
		t.Fatalf(".current-user length : expect '%v' got '%v' ", expected, got)
	}
	return db, server, user, err
}

// Expect is a helper used to reduce the boilerplate during test
func Expect(t *testing.T, got, want interface{}, comments ...string) {
	var comment string
	if want != got {
		if len(comments) > 0 {
			comment = comments[0]

		} else {
			comment = "Expect"
		}
		_, file, line, _ := runtime.Caller(1)
		t.Fatalf(fmt.Sprintf("Expect\r%s:%d:\r\t%s : %s", filepath.Base(file), line, comment, "want '%v' got '%v'."), want, got)
	}
}
