package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
	gonews "github.com/mparaiso/go-news/internal"
)

const version = "0.0.1-alpha"

func main() {

	startOptions := struct {
		Debug      bool
		Host, Port string
	}{}
	startFlagSet := flag.NewFlagSet("start", flag.ExitOnError)
	startFlagSet.BoolVar(&startOptions.Debug, "debug", false, "Start the application in Debug mode.")
	startFlagSet.StringVar(&startOptions.Host, "host", "localhost", "Host address of the server, example : localhost")
	startFlagSet.StringVar(&startOptions.Port, "port", "8080", "Server port, example: 8080")
	if len(os.Args) == 1 {
		print("go-news <command> [<options>]\n")
		print("go-news is a port of hacker-news. It is written in Go\n")
		print("commands: \n")
		print("\tstart start go-news server\n")
		print("\tversion prints the current version\n")
		print("\thelp prints the documentation\n")
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
		print("Usage : \n\n")
		print("go-news start [<options>] \n")
		startFlagSet.PrintDefaults()
		print("\n\texample: go-news start -debug -port 8080 -host localhost")
	default:
		print("not a valid command : ", os.Args[1])
		os.Exit(2)
	}

}
