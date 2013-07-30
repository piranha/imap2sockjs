package main

import (
	// "code.google.com/p/go-imap/go1/imap"
	"github.com/fzzy/sockjs-go/sockjs"
	"net/http"
	"log"
	)


func main() {
	static := http.FileServer(http.Dir("static/"))
	s := sockjs.NewServeMux(static)
	s.Handle("/api", connHandler, sockjs.NewConfig())

	println("starting on 7171...")
	err := http.ListenAndServe(":7171", s)
	if err != nil {
		log.Fatal(err)
	}
}

func connHandler(s sockjs.Session) {
	for {
		m := s.Receive()
		if m == nil {
			break
		}
		s.Send(m)
	}
}
