package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	gonews "github.com/mparaiso/go-news"
)

const version = "0.0.1-alpha"

func main() {
	defaultAddress := ":8080"
	flag.Parse()
	arguments := flag.Args()
	if len(arguments) > 0 {
		if flag.Arg(0) == "start" {
			options := gonews.DefaultContainerOptions()
			connection, connectionErr := sql.Open("sqlite3", "db.sqlite3")
			options.ConnectionFactory = func() (*sql.DB, error) {
				return connection, connectionErr
			}
			// start server
			app := gonews.GetApp(options, gonews.AppOptions{})
			fmt.Printf("Server Listening On: %s\n", defaultAddress)
			err := http.ListenAndServe(defaultAddress, app)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	fmt.Printf("version %s", version)
}
