// Integration tests

package gonews_test

import (
	"database/sql"
	"fmt"
	"time"

	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"

	"testing"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mparaiso/gonews/internal"
	"github.com/rubenv/sql-migrate"

	"flag"
	"net/url"
	"strings"
)

var DEBUG = false

// Allows arguments to be passed to test
// ex: go test -args -debug
func TestMain(m *testing.M) {
	debug := flag.Bool("debug", false, "debug the test suite")

	flag.Parse()
	DEBUG = *debug
	os.Exit(m.Run())
}

// Scenario: VISITING THE HOMEPAGE
// Given a server
// When / is requested
// It should return a valid response
// The correct number of threads should be displayed
func Test_Stories(t *testing.T) {
	db := GetDB(t)
	// Given a server
	server := GetServer(t, db)
	defer server.Close()

	// When the index is requested
	response, err := http.Get(server.URL + "/")

	if err != nil {
		t.Fatal(err)
	}
	// It should return a valid response
	if s := response.StatusCode; s != 200 {
		t.Fatalf("Status code should be 200, got %d", s)
	}
	doc, err := goquery.NewDocumentFromResponse(response)
	if err != nil {
		t.Fatal(err)
	}
	selection := doc.Find(".thread")

	// The correct number of threads should be displayed
	row := db.QueryRow("SELECT COUNT(id) FROM threads ;")
	var threadCount int
	err = row.Scan(&threadCount)
	if err != nil {
		t.Fatal(err)
	}
	if want, got := threadCount, selection.Length(); want != got {
		t.Fatalf(".threads length : want '%v' got '%v'", want, got)
	}
}

// Scenario: REQUESTING STORIES BY DOMAIN
// Given a server
// When /from?site=hipsters.acme is requested
// It should respond with status 200
// It should display the correct number of threads
func Test_Stories_By_Domain(t *testing.T) {
	// Given a server
	var err error
	site := "hipsters.acme"
	db := GetDB(t)
	server := GetServer(t, db)
	defer server.Close()
	// When /from?site=hipsters.acme is requested
	http.DefaultClient.Jar, err = cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.Get(server.URL + "/from?site=hipsters.acme")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	// It should respond with status 200
	if want, got := http.StatusOK, res.StatusCode; want != got {
		t.Fatalf("status code : want '%v' got '%v'", want, got)
	}
	row := db.QueryRow("SELECT COUNT(id) FROM threads WHERE url like ?", fmt.Sprintf("%%%s%%", site))
	var count int
	err = row.Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	// It should display the correct number of threads
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		t.Fatal(err)
	}
	selection := doc.Find(".thread")
	if want, got := count, selection.Length(); want != got {
		t.Fatalf(".thread length : want '%v' got '%v'", want, got)
	}
}

// Scenario: REQUESTING STORIES BY AUTHOR
// Given a server
// When /threads?id=1 is requested
// It should respond with status 200
// It should display the correct number of comments belonging to user with id 1
func Test_Comments_By_User(t *testing.T) {
	var err error
	db := GetDB(t)
	defer db.Close()
	// Given a server
	server := GetServer(t, db)
	defer server.Close()
	// When /threads?id=1 is requested
	id := 1
	res, err := http.Get(server.URL + fmt.Sprintf("/threads?id=%d", id))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	// It should respond with status 200
	if want, got := 200, res.StatusCode; want != got {
		t.Fatalf("status code : want '%v' got '%v' ", want, got)
	}
	// It should display the correct number of comments belonging to user with id 1
	row := db.QueryRow("SELECT COUNT(c.id) FROM comments c WHERE c.parent_id = ? AND c.author_id = ? LIMIT 1", 0, id)
	var count int
	err = row.Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		t.Fatal(err)
	}
	selection := doc.Find(".comments > .comment")
	if want, got := count, selection.Length(); want != got {
		t.Fatalf(".comments > .comment length : want '%v' got '%v' ", want, got)
	}
}

// Scenario:
// Given a server
// When /item?id=1 is requested
// It should respond with status 200
// The correct number of comments should be displayed
func Test_Story_By_ID(t *testing.T) {
	//	Given a server
	server := GetServer(t)
	defer server.Close()
	// When a thread with the id 1 is requested
	response, err := http.Get(server.URL + "/item?id=1")
	if err != nil {
		t.Fatal(err)
	}
	// It should respond with status 200
	if exp, got := 200, response.StatusCode; exp != got {
		t.Fatalf("Status code : want '%v' got '%v'", exp, got)
	}
	// The correct number of comments should be displayed
	doc, err := goquery.NewDocumentFromResponse(response)
	if err != nil {
		t.Fatal(err)
	}
	comments := doc.Find(".comment")

	if exp, got := 3, comments.Length(); exp != got {
		t.Fatalf(".comments .comment length  : want '%v' got '%v'", exp, got)
	}
}

// Scenario:
// Given a server
// When /item?id=3 is requested
// It should respond with status 200
// No comment should be displayed on the page
func Test_Story_By_ID_3(t *testing.T) {
	server := GetServer(t)
	defer server.Close()
	response, err := http.Get(server.URL + "/item?id=3")
	if err != nil {
		t.Fatal(err)
	}
	if exp, got := 200, response.StatusCode; exp != got {
		t.Fatalf("status code  : want '%v' got '%v'", exp, got)
	}
	doc, err := goquery.NewDocumentFromResponse(response)
	if err != nil {
		t.Fatal(err)
	}
	if expected, got := "0", doc.Find(".comment-count .count").Text(); expected != got {
		t.Fatalf("comment count  : want '%v' got '%v'", expected, got)
	}
}

// Scenario:
// Given a server
// When /user?id=1 is requested
// It should respond with status 200
// It should display the page for the user with id 1
func TestAppUserShow_1(t *testing.T) {
	server := GetServer(t)
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

// Given a server
// When /submitted?id=1 is requested
// It should respond with status 200
// It should display the list of stories submitted by that specific user
func TestAppSubmitted_id_1(t *testing.T) {
	server := GetServer(t)
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

// Scenario:
// Given a server
// When the login page is requested
// It should respond with status 200
// It should display the login form
func TestAppLogin_GET(t *testing.T) {
	server := GetServer(t)
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

// Scenario :
// Given a server
// When an unauthorized user attempts to visit a secured page
// It should respond with 401 error
func TestUnAuthorized(t *testing.T) {
	// Given a server
	server := GetServer(t)
	defer server.Close()

	// When an unauthorized user attempts to visit a secured page
	resp, err := http.Get(server.URL + "/submit")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// It should respond with 401 error
	if expected, got := 401, resp.StatusCode; expected != got {
		t.Fatalf("status code expected %v got %v", expected, got)
	}
}

// Scenario :
// Given a server
// When a user requests the login page
// It should respond with status 200
// When a user submits the login form with valid data
// It should login the user into the application
// It should redirect to the index with status 200
func TestAppLogin_POST(t *testing.T) {
	_, _, _, err := LoginUser(t)
	if err != nil {
		t.Fatal(err)
	}
}

// Scenario:
// Given a server
// When an authenicated user sends a post request to the logout page
// It should log the user out
// It should redirect to the index
func TestAppLogout(t *testing.T) {

	_, server, _, err := LoginUser(t)

	if err != nil {
		t.Fatal(err)
	}
	http.DefaultClient.Jar, err = cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	http.DefaultClient.CheckRedirect = nil
	res, err := http.Post(server.URL+"/logout", "", nil)
	defer res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if want, got := 200, res.StatusCode; want != got {
		t.Fatalf("status : want '%v' got '%v'", want, got)
	}
	if want, got := "/", res.Request.URL.Path; want != got {
		t.Fatalf("status : want '%v' got '%v'", want, got)
	}
}

// Scenario:
// Given a server
// When a user request the registration page
// It should respond with status 200
// When that user submits a registration form with correct values
// It should persists the new account into the database
// It should redirect the user to the login page
// It should create a new user in the database
func TestApp_Registration(t *testing.T) {
	var err error
	db := GetDB(t)
	server := GetServer(t, db)
	defer server.Close()
	http.DefaultClient.Jar, err = cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
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

	defer resp.Body.Close()

	if err != nil {
		t.Fatal(err)
	}

	if expected, got := 200, resp.StatusCode; expected != got {
		t.Fatalf("registration /login POST : status code : expected '%v' , got '%v'", expected, got)
	}
	if want, got := "/login", resp.Request.URL.Path; want != got {
		t.Fatalf("path: want '%v' got '%v' ", want, got)
	}
	row := db.QueryRow("SELECT username FROM users WHERE username = ? LIMIT 1", username)
	usernameResult := ""
	err = row.Scan(&usernameResult)
	if err != nil {
		t.Fatal(err)
	}
	if usernameResult != username {
		t.Fatalf("username : expected '%v' got '%v' ", username, usernameResult)
	}

}

// Scenario: REQUESTING AN NON EXISTING PAGE
// Given a server
// When a non existing page is requested
// It should respond with status 404
// The correct error message should be displayed
func TestApp_404(t *testing.T) {
	server := GetServer(t)
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
	errorMessage := selection.Find(".error-message").Text()
	if want, got := http.StatusText(http.StatusNotFound), errorMessage; want != got {
		t.Fatalf(".error-message text: want '%v' got '%v'", want, got)
	}
}

// Scenario: SUBMITTING A COMMENT
// Given a server
// When an authenticated client requests /item?id=1
// It should respond with status 200
// when an authenicated client submits a valid comment
// It should respond with status 200
// The number of comments on the story page should have increased by one

func Test_Server_Submitting_a_comment(t *testing.T) {

	// Given a server
	_, server, _, err := LoginUser(t)
	if err != nil {
		t.Fatal(err)
	}
	// When an authenticated client requests /item?id=1
	id := 1
	res, err := http.Get(fmt.Sprintf(server.URL+"/item?id=%d", id))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	// It should respond with status 200

	if want, got := 200, res.StatusCode; want != got {
		t.Fatalf("status : want '%v' got '%v' ", want, got)
	}
	doc, err := goquery.NewDocumentFromResponse(res)
	csrf, ok := doc.Find("input[name='comment_csrf']").First().Attr("value")
	if !ok {
		t.Fatalf("csrf value not found for comment form")
	}
	initialCommentNumber := doc.Find(".comment").Length()
	formValues := url.Values{
		"comment_content": {"this is a new comment"},
		"comment_csrf":    {csrf},
		"comment_submit":  {"submit"},
		"comment_parent":  {"0"},
		"comment_goto":    {fmt.Sprintf("/item?id=%d", id)},
		"comment_thread":  {fmt.Sprintf("%d", id)},
	}
	// when an authenicated client submits a valid comment
	res, err = http.Post(server.URL+"/comment", "application/x-www-form-urlencoded", strings.NewReader(formValues.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	// It should respond with status 200
	if want, got := 200, res.StatusCode; want != got {
		t.Fatalf("Status : want '%v' got '%v' ", want, got)
	}
	// The number of comments on the story page should have increased by one
	doc, err = goquery.NewDocumentFromResponse(res)
	if err != nil {
		t.Fatal(err)
	}
	if want, got := initialCommentNumber, doc.Find(".comment").Length(); want != got {
		t.Fatalf(".comment length : want '%v' got '%v' ", want, got)
	}
}

// Scenario:
// Given a server
// When an authenticated user requests the submission page
// It should respond with status 200
// It should display the story submission form
// When an authenticated user submits a valid story submission
// It should create a new Thread in the database
// It should create a thread vote with the id of the thread and the id of the author
// It should redirect to the story page with the right ID
// It should respond with status 200
// It should display the right story
func TestSubmitStory(t *testing.T) {
	// Given a server
	db, server, user, err := LoginUser(t)
	defer server.Close()

	if err != nil {
		t.Fatal(err)
	}
	// When an authenticated user requests the submission page
	res, err := http.Get(server.URL + "/submit")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	// It should respond with status 200
	if want, got := 200, res.StatusCode; want != got {
		t.Fatalf("status : want '%v' got '%v' ", want, got)
	}
	// It should display the story submission form
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		t.Fatal(err)
	}
	selection := doc.Find("form[name='submission']")
	if want, got := 1, selection.Length(); want != got {
		t.Fatalf("form[name='submission'] length : want '%v' got '%v' ", want, got)
	}
	// When an authenticated user submits a valid story submission
	csrf, _ := doc.Find("#submission_csrf").Attr("value")
	submissionForm := &gonews.SubmissionForm{Title: "Serverless development on Amazon AWS with Opex", CSRF: csrf, URL: "http://presentation.opex.com/index.html?foobar=biz#baz"}
	values := url.Values{
		"submission_title": {submissionForm.Title},
		"submission_csrf":  {submissionForm.CSRF},
		"submission_url":   {submissionForm.URL},
	}

	res, err = http.Post(server.URL+"/submit", "application/x-www-form-urlencoded", strings.NewReader(values.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	// It should respond with status 200
	if want, got := 200, res.StatusCode; want != got {
		t.Fatalf("status : want '%v' got '%v' ", want, got)
	}
	// It should create a new Thread in the database
	row := db.QueryRow("SELECT threads.id,threads.title from threads where threads.title = ?  AND threads.author_id;", submissionForm.Title, user.ID)
	var title string
	var id int64
	err = row.Scan(&id, &title)
	if err != nil {
		t.Fatal(err)
	}
	if want, got := values.Get("submission_title"), title; want != got {
		t.Fatalf("story title : want '%v' got '%v' ", want, got)
	}
	// It should create a thread vote with the id of the thread and the id of the author
	row = db.QueryRow("SELECT tv.id FROM thread_votes tv where tv.author_id = ? and tv.thread_id = ? ", user.ID, id)
	var threadVoteID int64

	err = row.Scan(&threadVoteID)
	if err != nil {
		t.Fatal(err)
	}

	// It should redirect to the story page with the right ID
	if want, got := fmt.Sprintf("%s/item?id=%d", server.URL, id), res.Request.URL.String(); want != got {
		t.Fatalf("redirection path : want '%v' got '%v'", want, got)
	}

	// It should display the right story
	doc, err = goquery.NewDocumentFromResponse(res)
	if err != nil {
		t.Fatal(err)
	}

	if want, got := title, doc.Find(".thread-title").First().Text(); want != got {
		t.Fatalf("story title : want '%v' got '%v' ", want, got)
	}
}

// Directory is the current directory
var Directory = func() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return dir
}()

// GetDB gets the db connection
func GetDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	return db
}

// MigrateUp executes db migrations
func MigrateUp(db *sql.DB, t *testing.T) *sql.DB {
	_, err := migrate.Exec(db, "sqlite3", migrate.FileMigrationSource{"./../migrations/development/sqlite3"}, migrate.Up)
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
	app := gonews.GetApp(gonews.AppOptions{ContainerOptions: GetContainerOptions(db)})
	server := httptest.NewServer(app)

	logger := &log.Logger{}
	logger.SetOutput(os.Stdout)
	server.Config.ErrorLog = logger
	server.Config.WriteTimeout = 3 * time.Second
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
