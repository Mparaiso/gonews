// @Copyright (c) 2016 mparaiso <mparaiso@online.fr>  All rights reserved.

package gonews_test

import (
	"database/sql"
	"time"

	"log"
	"net/http"
	"net/http/httptest"
	"os"
	// "path"
	"testing"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mparaiso/go-news/internal"
	"github.com/rubenv/sql-migrate"

	"net/url"
	"strings"
	"sync"
)

var DEBUG = false

func TestAppIndex(t *testing.T) {
	server := SetUp(t)
	defer server.Close()
	response, err := http.Get(server.URL + "/")

	// Test
	if err != nil {
		t.Fatal(err)
	}
	if s := response.StatusCode; s != 200 {
		t.Fatalf(StatusShouldF, s)
	}
	doc, err := goquery.NewDocumentFromResponse(response)
	if err != nil {
		t.Fatal(err)
	}
	selection := doc.Find(".thread")
	Log(doc.Html())
	if expected, got := 6, selection.Length(); expected != got {
		t.Fatalf(LabelExpectGotF, ".threads length", expected, got)
	}
}

func TestAppThreadShow(t *testing.T) {
	server := SetUp(t)
	defer server.Close()
	response, err := http.Get(server.URL + "/thread?id=1")
	if err != nil {
		t.Fatal(err)
	}
	if exp, got := 200, response.StatusCode; exp != got {
		t.Fatalf(LabelExpectGotF, "Status code", exp, got)
	}
	doc, err := goquery.NewDocumentFromResponse(response)
	if err != nil {
		t.Fatal(err)
	}
	comments := doc.Find(".comment")
	Log(doc.Html())
	if exp, got := 3, comments.Length(); exp != got {
		t.Fatalf(LabelExpectGotF, ".comments .comment length", exp, got)
	}
}

func TestAppThreadShow_with_no_comment(t *testing.T) {
	server := SetUp(t)
	defer server.Close()
	response, err := http.Get(server.URL + "/thread?id=3")
	if err != nil {
		t.Fatal(err)
	}
	if exp, got := 200, response.StatusCode; exp != got {
		t.Fatalf(LabelExpectGotF, "status code", exp, got)
	}
	doc, err := goquery.NewDocumentFromResponse(response)
	if err != nil {
		t.Fatal(err)
	}
	if expected, got := "0", doc.Find(".comment-count .count").Text(); expected != got {
		t.Fatalf(LabelExpectGotF, "comment count", expected, got)
	}
}

func TestAppUserShow_1(t *testing.T) {
	server := SetUp(t)
	defer server.Close()
	res, err := http.Get(server.URL + "/user?id=1")
	if err != nil {
		t.Fatal(err)
	}
	if expected, got := 200, res.StatusCode; expected != got {
		t.Fatalf("status code : expect '%v' ,got '%v' ", expected, got)
	}
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		t.Fatal(err)
	}
	if expected, got := "johndoe", doc.Find(".username").First().Text(); expected != got {
		t.Fatalf(".user text : expect '%v' , got '%v'", expected, got)
	}
}

func TestAppSubmitted_id_1(t *testing.T) {
	server := SetUp(t)
	defer server.Close()
	resp, err := http.Get(server.URL + "/submitted?id=1")
	if err != nil {
		t.Fatal(err)
	}
	if expected, got := 200, resp.StatusCode; expected != got {
		t.Fatalf("StatusCode: expected '%v', got '%v'", expected, got)
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		t.Fatal(err)
	}
	selection := doc.Find(".thread")
	if expected, got := 2, selection.Length(); expected != got {
		t.Fatalf(".thread length: expected '%v', got '%v'", expected, got)
	}

}

func TestAppLogin_GET(t *testing.T) {
	server := SetUp(t)
	defer server.Close()
	resp, err := http.Get(server.URL + "/login")
	if err != nil {
		t.Fatal(err)
	}
	if expected, got := 200, resp.StatusCode; expected != got {
		t.Fatalf("status code: expected '%v' got '%v'", expected, got)
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		t.Fatal(err)
	}
	selection := doc.Find(`form[name='login']`)
	if expected, got := 1, selection.Length(); expected != got {
		t.Fatalf("form[name=login] length: expected '%v' got '%v'", expected, got)
	}
	selection = doc.Find(`form[name='registration']`)
	if expected, got := 1, selection.Length(); expected != got {
		t.Fatalf("form[name=login] length: expected '%v' got '%v'", expected, got)
	}
}

// TestAppLogin_POST logs a registered user into the application
func TestAppLogin_POST(t *testing.T) {
	_, _, _, err := LoginUserHelper(t)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAppLogout(t *testing.T) {
	// db, server, user, err := LoginUserHelper(t)

}

// TestAppLogin_POST_registration tests the registration process and verifies
// the new user has been persisted into the db
func TestApp_Registration(t *testing.T) {
	db := GetDB(t)
	server := SetUp(t, db)
	defer server.Close()
	http.DefaultClient.Jar = NewTestCookieJar()
	resp, err := http.Get(server.URL + "/login")
	defer resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if expected, got := 200, resp.StatusCode; expected != got {
		t.Fatalf("status : expected '%v', got '%v'", expected, got)
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		t.Fatal(err)
	}
	registrationCsrf, exists := doc.Find("input[name='registration_csrf']").First().Attr("value")
	if !exists {
		t.Fatal("registration csrf value not found")
	}
	username := "jefferson"
	values := url.Values(map[string][]string{
		"registration_csrf":                  {registrationCsrf},
		"registration_username":              {username},
		"registration_password":              {"password"},
		"registration_password_confirmation": {"password"},
		"registration_email":                 {"jefferson@acme.com"},
	})
	resp, err = http.Post(server.URL+"/register", "application/x-www-form-urlencoded", strings.NewReader(values.Encode()))
	// defer resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if expected, got := 200, resp.StatusCode; expected != 200 {
		t.Fatalf("registration /login POST : status code : expected '%v' , got '%v'", expected, got)
	}
	// check db if record was created
	row := db.QueryRow("SELECT username FROM users WHERE username = ? LIMIT 1", username)
	usernameResult := ""
	err = row.Scan(&usernameResult)
	if err != nil {
		t.Fatal(err)
	}
	if usernameResult != username {
		t.Fatalf("username : expected '%v' got '%v' ", username, usernameResult)
	}
	///http.CookieJar
}

// LoginUserHelper logs a user before executing a test
func LoginUserHelper(t *testing.T) (*sql.DB, *httptest.Server, *gonews.User, error) {
	// Setup
	db := GetDB(t)
	server := SetUp(t, db)
	unencryptedPassword := "password"
	user := &gonews.User{Username: "mike_doe", Email: "mike_doe@acme.com"}
	user.CreateSecurePassword(unencryptedPassword)
	result, err := db.Exec("INSERT INTO users(username,email,password) values(?,?,?);", user.Username, user.Email, user.Password)
	if err != nil {
		t.Fatal(err)
	}
	// t.Logf("%#v", user)
	if n, err := result.RowsAffected(); err != nil || n != 1 {
		t.Fatal(n, err)
	}
	defer server.Close()
	http.DefaultClient.Jar = NewTestCookieJar()
	http.DefaultClient.CheckRedirect = nil
	// test
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

	if expected, got := 301, res.StatusCode; expected != got {
		//t.Logf(" %s %s", ioutil.ReadAll(res.Body))
		t.Fatalf("POST /login status : expected '%v' got '%v'", expected, got)
	}
	doc, err = goquery.NewDocumentFromResponse(res)

	t.Log(doc.Html())

	if err != nil {
		t.Fatal(err)
	}
	selection = doc.Find(".current-user")
	if expected, got := 1, selection.Length(); expected != got {
		t.Fatalf(".current-user length : expect '%v' got '%v' ", expected, got)
	}
	return db, server, user, err
}

func TestApp_404(t *testing.T) {
	server := SetUp(t)
	defer server.Close()
	resp, err := http.Get(server.URL + "/non-existant-route")
	if err != nil {
		t.Fatal(err)
	}
	if exp, got := 404, resp.StatusCode; exp != got {
		t.Fatalf("Status code expected %v got %v ", exp, got)
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		t.Fatal(err)
	}
	selection := doc.Find(".error")
	if expected, got := 1, selection.Length(); expected != got {
		t.Fatalf(".error length: expected '%v', got '%v'", expected, got)
	}
}

const StatusShouldF = "Status code should be 200, got %d"

const LabelExpectGotF = "'%s': expected %v, got %v"

// Directory is the current directory
var Directory = func() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return dir
}()

// var MigrationDirectory = path.Join(path.Clean(Directory), "..", "migrations", "development", "sqlite3")

func GetDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func MigrateUp(db *sql.DB, t *testing.T) *sql.DB {
	_, err := migrate.Exec(db, "sqlite3", migrate.FileMigrationSource{"./../migrations/development/sqlite3"}, migrate.Up)
	if err != nil {
		t.Fatal(err)
	}
	return db
}
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

func SetUp(t *testing.T, dbs ...*sql.DB) *httptest.Server {
	// Set Up
	var db *sql.DB
	if len(dbs) == 0 {
		db = GetDB(t)
	} else {
		db = dbs[0]
	}
	MigrateUp(db, t)
	app := gonews.GetApp(gonews.AppOptions{ContainerOptions: GetContainerOptions(db)})
	server := httptest.NewServer(app)
	logger := &log.Logger{}
	logger.SetOutput(os.Stdout)
	server.Config.ErrorLog = logger
	server.Config.WriteTimeout = 3 * time.Second
	return server
}
func Log(args ...interface{}) {
	if DEBUG {
		log.Print(args...)
	}
}

type TestCookieJar struct {
	cookieJar map[string][]*http.Cookie
	mutex     *sync.RWMutex
}

func NewTestCookieJar() *TestCookieJar {
	return &TestCookieJar{mutex: &sync.RWMutex{}, cookieJar: make(map[string][]*http.Cookie)}
}
func (jar TestCookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.mutex.Lock()
	jar.cookieJar[u.Host] = cookies
	jar.mutex.Unlock()
}
func (jar TestCookieJar) Cookies(u *url.URL) (cookies []*http.Cookie) {

	jar.mutex.Lock()
	cookies = jar.cookieJar[u.Host]
	jar.mutex.Unlock()
	return
}
