package gonews_test

import "testing"
import gonews "github.com/mparaiso/gonews/core"

func TestThreadRepository_GetByAuthorID(t *testing.T) {
	db := MigrateUp(GetDB(t), t)
	threadRepository := &gonews.ThreadRepository{DB: db, Logger: gonews.NewDefaultLogger(gonews.OFF)}
	threads, err := threadRepository.GetByAuthorID(1)
	if err != nil {
		t.Fatal(err)
	}
	if expected, got := 2, len(threads); expected != got {
		t.Fatalf("threads length: expected '%v' , got '%v'", expected, got)
	}
	// t.Logf("%#v %#v", threads[0], threads[1])
	if expected, got := int64(1), threads[0].ID; expected != got {
		t.Fatalf("threads[0].ID : expected '%v' , got '%v' ", expected, got)
	}
	if expected, got := int64(1), threads[0].AuthorID; expected != got {
		t.Fatalf("threads[0].AuthorID: expected '%v' , got '%v'", expected, got)
	}
}
