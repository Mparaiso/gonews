package main

import (
	"flag"
	"fmt"
	"github.com/mparaiso/hn-go"
	"log"
	"net/http"
)

func main() {
	defaultAddress := ":8080"
	flag.Parse()
	arguments := flag.Args()
	if len(arguments) > 0 {
		if flag.Arg(0) == "start" {
			// start server
			server := hn.GetServer()
			err := http.ListenAndServe(defaultAddress, server)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	fmt.Print("HN-GO! version 0.01-alpha")
}
