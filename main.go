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

package main

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

	_ "github.com/mattn/go-sqlite3"
	gonews "github.com/mparaiso/gonews/core"
	sqlmigrate "github.com/rubenv/sql-migrate"

	"gopkg.in/yaml.v2"
)

// Version is the gonews server version
var Version = "0.0.1-alpha"

const defaultSecret = "please change this key in production!"

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

	startOptions, startFlagSet := DeclareStartOptions()

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
		// configuration

		startFlagSet.Parse(os.Args[2:])

		// if port is an env variable, get env variable
		if len(startOptions.Port) > 0 && startOptions.Port[0] == '$' {
			startOptions.Port = os.Getenv(string(startOptions.Port[1:]))
		}

		connection, connectionErr := sql.Open(startOptions.Driver, startOptions.DataSource)
		if connectionErr != nil {
			log.Fatal(connectionErr)
		}

		// warnings
		if startOptions.Env != "production" {
			log.Printf("You are using '%s' environment, please use option -env=production in production", startOptions.Env)
		}
		if startOptions.Secret == defaultSecret {
			log.Printf("You are using the default secret key which is unsecure, please generate a strong secret key, and set is with -secret argument")
		}
		// migration
		if startOptions.Migrate {
			migrationSource := sqlmigrate.FileMigrationSource{Dir: path.Join(startOptions.MigrationPath, startOptions.Env, startOptions.Driver)}
			i, err := sqlmigrate.Exec(connection, startOptions.Driver, migrationSource, sqlmigrate.Up)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("%d migrations executed", i)
		}
		// loading fixtures
		if startOptions.LoadFixtures {
			log.Println("loading fixtures ... ")
			LoadFixtures(connection, startOptions.Driver)
			log.Println("done loading fixtures.")
		}
		// start server
		containerOptions := gonews.DefaultContainerOptions()
		containerOptions.LogLevel = gonews.LogLevel(startOptions.LogLevel)
		containerOptions.Debug = startOptions.Debug
		containerOptions.ConnectionFactory = func() (*sql.DB, error) {
			return connection, connectionErr
		}
		appOptions := gonews.AppOptions{ContainerOptions: containerOptions}
		// configuration file
		if startOptions.ConfigurationFilePath != "" {
			fileBytes, err := ioutil.ReadFile(startOptions.ConfigurationFilePath)
			if err != nil && startOptions.ConfigurationFilePath != "gonews.yml" {
				log.Printf("Error loading configuration file %s : %s \n", startOptions.ConfigurationFilePath, err)
			} else {
				err = yaml.Unmarshal(fileBytes, &appOptions)
				if err != nil {
					log.Fatalf("Error reading configuration file %s : %s \n", startOptions.ConfigurationFilePath, err)
				} else {
					log.Printf("Config loaded : %#v", appOptions)
				}
			}
		}
		app := gonews.GetApp(appOptions)
		addr := startOptions.Host + ":" + startOptions.Port
		fmt.Printf("Server Listening On: %s\n", addr)
		err := http.ListenAndServe(addr, app)
		if err != nil {
			log.Fatal(err)
		}
		return
	case "version":
		print(Version)
	case "help":
		printDocumentation()
	default:
		print("not a valid command : ", os.Args[1])
		printDocumentation()
		os.Exit(2)
	}

}

// DeclareStartOptions declare start options to be parsed from command line arguments
func DeclareStartOptions() (*StartOptions, *flag.FlagSet) {
	startOptions := &StartOptions{}
	startFlagSet := flag.NewFlagSet("start", flag.ExitOnError)
	startFlagSet.StringVar(&startOptions.ConfigurationFilePath, "config", "gonews.yml", "Configuration file path")
	startFlagSet.BoolVar(&startOptions.Debug, "debug", false, "Starts the application in Debug mode.")
	startFlagSet.BoolVar(&startOptions.LoadFixtures, "loadfixtures", false, "Load sample datas into the database, should obviously be executed only once.")
	startFlagSet.StringVar(&startOptions.Host, "host", "0.0.0.0", "Host address of the server, example : localhost")
	startFlagSet.StringVar(&startOptions.Port, "port", "8080", "Server port, example: 8080")
	startFlagSet.StringVar(&startOptions.Env, "env", "development", "Current environment, examples: -env=developement , -env=test ")
	startFlagSet.StringVar(&startOptions.Secret, "secret", defaultSecret, "Secret key used for encryption, example -secret=\"my-secret-key\"")
	startFlagSet.BoolVar(&startOptions.Migrate, "migrate", false, "migrate will execute an upward database migration when the application starts")
	startFlagSet.StringVar(&startOptions.MigrationPath, "migrationpath", "migrations", "Sets the migration path from where migrations are executed")
	startFlagSet.StringVar(&startOptions.Driver, "driver", "sqlite3", "Sets the database driver. Example : -driver=sqlite3")
	startFlagSet.StringVar(&startOptions.DataSource, "datasource", "db.sqlite3", "Sets the datasource. Example: -datasource=db.sqlite3")
	startFlagSet.IntVar(&startOptions.LogLevel, "loglevel", 1, "A value between 0 and 6. Sets the logger verbosity level. Example: -loglevel 0 ")

	return startOptions, startFlagSet
}

// StartOptions are arguments passed to the commandline
// when start action is invoked
type StartOptions struct {
	Debug, Migrate, LoadFixtures bool
	Host, Port,
	Env, Driver,
	DataSource, MigrationPath,
	ConfigurationFilePath,
	Secret string
	LogLevel int
}

// LoadFixtures loads test fixtures in a transaction
// if the transaction fails, no fatal error is triggered
func LoadFixtures(db *sql.DB, driver string) *sql.DB {

	migrationFile, err := os.Open("./testdata/fixtures/" + driver + "/fixtures.sql")
	if err != nil {
		log.Fatal(err)
	}
	buffer := new(bytes.Buffer)
	_, err = buffer.ReadFrom(migrationFile)
	if err != nil {
		log.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	_, err = tx.Exec(buffer.String())
	if err != nil {
		log.Printf("Error loading fixtures, %s", err)
		tx.Rollback()
	} else {
		err = tx.Commit()
		if err != nil {
			log.Fatal(err)
		}
	}
	return db
}
