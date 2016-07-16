package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
	gonews "github.com/mparaiso/gonews/core"
)

const version = "0.0.1-alpha"

const documentation = `
gonews is a port of hacker-news. It is written in Go

Usage:
go-news <command> [<options>]

Commands: 
	start 	Starts go-news server
	version Prints the current version
	help 	Prints the documentation
`

func main() {

	startOptions := struct {
		Debug      bool
		Host, Port string
	}{}
	startFlagSet := flag.NewFlagSet("start", flag.ExitOnError)
	startFlagSet.BoolVar(&startOptions.Debug, "debug", false, "Starts the application in Debug mode.")
	startFlagSet.StringVar(&startOptions.Host, "host", "localhost", "Host address of the server, example : localhost")
	startFlagSet.StringVar(&startOptions.Port, "port", "8080", "Server port, example: 8080")
	printDocumentation := func() {
		print(documentation)
		print("\nstart command options :\n\n")
		startFlagSet.PrintDefaults()
		print("\nexample: gonews start -debug -port 8080 -host localhost\n")
	}
	if len(os.Args) == 1 {
		printDocumentation()
		return
	}
	switch os.Args[1] {
	case "start":
		startFlagSet.Parse(os.Args[2:])
		options := gonews.DefaultContainerOptions()
		options.Debug = startOptions.Debug
		connection, connectionErr := sql.Open("sqlite3", "db.sqlite3")
		options.ConnectionFactory = func() (*sql.DB, error) {
			return connection, connectionErr
		}
		// start server
		app := gonews.GetApp(gonews.AppOptions{ContainerOptions: options})
		addr := startOptions.Host + ":" + startOptions.Port
		fmt.Printf("Server Listening On: %s\n", addr)
		err := http.ListenAndServe(addr, app)
		if err != nil {
			log.Fatal(err)
		}
		return
	case "version":
		print(version)
	case "help":
		printDocumentation()
	default:
		print("not a valid command : ", os.Args[1])
		printDocumentation()
		os.Exit(2)
	}

}
