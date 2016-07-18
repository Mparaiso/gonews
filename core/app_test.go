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
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/mparaiso/gonews/core"
)

// Scenario: VISITING THE HOMEPAGE
// Given a server
// When / is requested
// It should return a valid response
// The correct number of threads should be displayed
func Test_Visiting_the_homepage(t *testing.T) {
	db := GetDB(t)
	// Given a server
	server := GetServer(t, db)
	defer server.Close()

	// When the index is requested
	response, err := http.Get(server.URL + gonews.Route{}.StoriesByScore())
	Expect(t, err, nil)

	// It should return a valid response
	Expect(t, response.StatusCode, 200, "status")
	doc, err := goquery.NewDocumentFromResponse(response)
	Expect(t, err, nil)
	selection := doc.Find(".thread")

	// The correct number of threads should be displayed
	row := db.QueryRow("SELECT COUNT(id) FROM threads ;")
	var threadCount int
	err = row.Scan(&threadCount)
	Expect(t, err, nil)
	Expect(t, selection.Length(), threadCount, ".threads length")
}

// Scenario: REQUESTING STORIES BY DOMAIN
// Given a server
// When /from?site=hipsters.acme is requested
// It should respond with status 200
// It should display the correct number of threads
func TestRequestingStoriesByDomain(t *testing.T) {
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
	res, err := http.Get(server.URL + gonews.Route{}.StoriesByDomain() + "?site=hipsters.acme")
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

// Scenario: REQUESTING COMMENTS BY AUTHOR
// Given a server
// When /threads?id=1 is requested
// It should respond with status 200
// It should display the correct number of comments belonging to user with id 1
func TestRequestingCommentsByUser(t *testing.T) {
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

// Scenario: REQUESTING A STORY BY ID
// Given a server
// When /item?id=1 is requested
// It should respond with status 200
// The correct number of comments should be displayed
func TestRequestingAStoryByID(t *testing.T) {
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

// Scenario: REQUESTING A STORY PAGE BY ID 3
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

// Scenario: REQUESTING A USER PROFILE PAGE BY ID
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

// Given a server REQUESTING STORIES BY AUTHOR
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

// Scenario: REQUESTING THE LOGIN PAGE
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

// Scenario : REQUESTING AN UNAUTHORIZED PAGE
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

// Scenario : SIGNING INTO THE APPLICATION
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

// Scenario: SIGNING A USER OUT
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

// Scenario: REGISTERING A NEW ACCOUNT
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

func TestSubmittingAComment(t *testing.T) {

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
		"comment_content":   {"this is a new comment"},
		"comment_csrf":      {csrf},
		"comment_submit":    {"submit"},
		"comment_parent_id": {"0"},
		"comment_goto":      {fmt.Sprintf("/item?id=%d", id)},
		"comment_thread_id": {fmt.Sprintf("%d", id)},
	}
	// when an authenicated client submits a valid comment
	res, err = http.Post(server.URL+gonews.Route{}.Reply(), "application/x-www-form-urlencoded", strings.NewReader(formValues.Encode()))
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
	if want, got := initialCommentNumber+1, doc.Find(".comment").Length(); want != got {
		t.Fatalf(".comment length : want '%v' got '%v' ", want, got)
	}
}

// Scenario: SUBMITTING A NEW STORY
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
	defer func() { db.Close(); server.Close() }()

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

// Scenario: REPLYING TO COMMENT
// Given a server
// When an authenticated user requests the "reply to comment" page /reply?id=XXX&goto=/item?id=XXXX
// it should respond with status 200
// When a valid comment form is submitted
// it should respond with status 200
// it should redirect to initial requested page
func TestReplyingToComment(t *testing.T) {
	// Given a server
	db, server, _, err := LoginUser(t)
	defer func() {
		db.Close()
		server.Close()
	}()
	Expect(t, err, nil)
	Expect(t, err, nil)
	threadID := 1
	res, err := http.Get(fmt.Sprintf("%s/item?id=%d", server.URL, threadID))
	Expect(t, err, nil)
	defer res.Body.Close()
	Expect(t, res.StatusCode, 200, "status")
	doc, err := goquery.NewDocumentFromResponse(res)
	Expect(t, err, nil)
	href, ok := doc.Find(".comment-reply").First().Attr("href")
	Expect(t, ok, true, "first .comment-reply href not found")
	// When an authenticated user requests the "reply to comment" page /reply?id=XXX&goto=/item?id=XXXX
	res, err = http.Get(server.URL + href)
	Expect(t, err, nil)
	defer res.Body.Close()
	// it should respond with status 200
	Expect(t, res.StatusCode, 200, "status")
	doc, err = goquery.NewDocumentFromResponse(res)
	Expect(t, err, nil)
	csrf, ok := doc.Find("input[name='comment_csrf']").First().Attr("value")
	Expect(t, ok, true, "first input[name='comment_csrf'] value not found")
	parentID, ok := doc.Find("input[name='comment_parent_id']").First().Attr("value")
	if !ok {
		t.Fatalf("first input[name='comment_parent_id'] value not found")
	}
	formThreadID, ok := doc.Find("input[name='comment_thread_id']").First().Attr("value")
	if !ok {
		t.Fatalf("first input[name='comment_thread_id'] value not found")
	}
	Goto, ok := doc.Find("input[name='comment_goto']").First().Attr("value")
	if !ok {
		t.Fatalf("first input[name='comment_goto'] value not found")
	}
	formValues := url.Values{
		"comment_content":   {"this is a response to a comment"},
		"comment_csrf":      {csrf},
		"comment_submit":    {"submit"},
		"comment_parent_id": {parentID},
		"comment_goto":      {Goto},
		"comment_thread_id": {formThreadID},
	}
	action, ok := doc.Find("form[name='comment']").First().Attr("action")
	if !ok {
		t.Fatalf("first form[name='comment'] action attribute not found")
	}
	// When a valid comment form is submitted
	res, err = http.Post(server.URL+action, "application/x-www-form-urlencoded", strings.NewReader(formValues.Encode()))
	Expect(t, err, nil)
	defer res.Body.Close()
	// it should respond with status 200
	Expect(t, res.StatusCode, 200, "status")
	row := db.QueryRow("SELECT ID from comments_view ORDER BY Created DESC,ID DESC LIMIT 1")
	var id int
	err = row.Scan(&id)
	Expect(t, err, nil, "row scan")
	// it should redirect to initial requested page
	Expect(t, res.Request.URL.RequestURI()+"#"+res.Request.URL.Fragment, fmt.Sprintf("%s#%d", Goto, id), "location")
}

// Scenario: REQUESTING NEW COMMENTS PAGE
// Given a server
// When the /newcomments url is requested
// It should respond with status 200
// It should display the right number of comments
// It should display the newest comment first
func TestRequestingNewCommentsPage(t *testing.T) {
	http.DefaultClient.Jar = nil
	db := GetDB(t)
	server := GetServer(t, db)
	defer func() {
		db.Close()
		server.Close()
	}()
	// When the /newcomments url is requested
	res, err := http.Get(server.URL + gonews.Route{}.NewComments())
	// It should respond with status 200
	Expect(t, err, nil, "error")
	Expect(t, res.StatusCode, 200, "status")
	// It should display the right number of comments
	row := db.QueryRow("SELECT COUNT(id) FROM comments ;")
	var count int
	Expect(t, row.Scan(&count), nil)
	doc, err := goquery.NewDocumentFromResponse(res)
	Expect(t, err, nil)
	Expect(t, doc.Find(".comment").Length(), count, ".comment count")
	// It should display the newest comment first
	row = db.QueryRow("SELECT content FROM comments ORDER BY created DESC lIMIT 1")
	var content string
	Expect(t, row.Scan(&content), nil)
	Expect(t, doc.Find(".comment > .content").First().Text(), content, ".comment > . content text")
}

// Scenario: DISPLAYING NEWEST STORIES PAGE
// Given a server
// When the /newest url is requested
// It should respond with status 200
// It should display the newest story first
func TestDisplayingNewestStoriesPage(t *testing.T) {
	db := GetDB(t)
	// Given a server
	server := GetServer(t, db)
	defer func() {
		db.Close()
		server.Close()
	}()
	http.DefaultClient.Jar = nil
	// When the /newest url is requested
	res, err := http.Get(server.URL + gonews.Route{}.NewStories())
	Expect(t, err, nil, "GET /newest")
	// It should respond with status 200

	Expect(t, res.StatusCode, 200, "status")
	// It should display the newest story first
	var id int64
	row := db.QueryRow("SELECT id FROM threads ORDER BY created DESC LIMIT 1")
	err = row.Scan(&id)
	Expect(t, err, nil, "thread id")
	doc, err := goquery.NewDocumentFromResponse(res)
	Expect(t, err, nil, "selection")
	commentID := doc.Find(".thread").First().AttrOr("data-thread-id", "")
	// It should display the newest story first
	Expect(t, commentID, fmt.Sprintf("%d", id), ".thread[data-thread-id]")
}

// Scenario: UPVOTING A STORY
// Given a server
// When an authenticated user requests the homepage
// Then the authenticated user upvotes the first story he can vote on
// It should respond with status 200
// It should returns to the page the vote was casted from
// The user should have created a new thread_vote
// TODO
//func TestUpvotingAStory(t *testing.T) {
//	t.Log("UPVOTING A STORY")
//	t.Log("Given a server")
//	db, server, user, err := LoginUser(t)
//	defer func() {
//		db.Close()
//		server.Close()
//	}()
//	t.Log("When an authenticated user requests the homepage")
//	res, err := http.Get(server.URL + gonews.Route{}.StoriesByScore())
//	Expect(t, err, nil)
//	Expect(t, res.StatusCode, 200, "\tstatus")
//	/* form should take the form off
//	<form action="/vote/thread" method="post" name="thread_vote">
//		<input type="hidden" name="thread_vote_thread_id" value=""/>
//		<input type="hidden" name="thread_vote_goto" value="/?item=34949#thread=349349"/>
//		<input type submit value="&utrif;" name="thread_vote_submit" />
//	</form>
//	*/
//	t.Log("Then the authenticated user upvotes the first story he can vote on")
//	doc, err := goquery.NewDocumentFromResponse(res)
//	Expect(t, err, nil)
//	form := doc.Find("form[name='thread_vote']").First()
//	Expect(t, form.Length(), 1, "form[name='thread_vote'] length")

//	values := url.Values{
//		"thread_vote_thread_id": {form.Find("input[name='thread_vote_thread_id]").First().AttrOr("value", "")},
//		"thread_vote_goto":      {form.Find("input[name='thread_vote_goto]").First().AttrOr("value", "")},
//		"thread_vote_submit":    {form.Find("input[name='thread_vote_submit]").First().AttrOr("value", "")},
//	}
//	row := db.QueryRow("SELECT count(id) FROM thread_votes WHERE thread_votes.author_id = ? ", user.ID)
//	var threadVoteCount int
//	err = row.Scan(&threadVoteCount)
//	Expect(t, err, nil)
//	res, err = http.Post(server.URL+gonews.Route{}.CastStoryVote(), FORM_MIME_TYPE, strings.NewReader(values.Encode()))
//	Expect(t, err, nil)
//	t.Log("It should respond with status 200")
//	Expect(t, res.StatusCode, 200, "status")
//	location, err := res.Location()
//	Expect(t, err, nil)
//	t.Log("It should returns to the page the vote was casted from")
//	Expect(t, location.RequestURI(), gonews.Route{}.StoriesByScore(), "location")
//	var newThreadVoteCount int
//	row = db.QueryRow("SELECT count(id) FROM thread_votes WHERE thread_votes.author_id = ? ", user.ID)
//	err = row.Scan(&newThreadVoteCount)
//	Expect(t, err, nil)
//	t.Log("The user should have created a new thread_vote")
//	Expect(t, newThreadVoteCount, threadVoteCount+1, "thread_votes count")
//}
