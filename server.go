package hn

import "net/http"

func GetServer() http.Handler {
	server := http.NewServeMux()
	server.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("Hello HN-Go!"))
	})
	return server
}
