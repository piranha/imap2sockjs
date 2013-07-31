package main

import (
	// "code.google.com/p/go-imap/go1/imap"
	"github.com/fzzy/sockjs-go/sockjs"
	"net/http"
	"log"
	"encoding/json"
	"fmt"
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
	c, err := DialTLS("imap.gmail.com", nil)
	if err != nil {
		log.Fatal(err)
		s.Send(err2byte(err))
		s.End()
		return
	}

	go func() {
		for {
			r, err := c.Receive()
			if err != nil {
				s.Send(err2byte(err))
				continue
			}
			b, err := json.Marshal(r)
			if err != nil {
				s.Send(err2byte(err))
				continue
			}
			s.Send(b)
		}
	}()

	for {
		m := s.Receive()
		if m == nil {
			break
		}
		cmd := &Command{}
		err := json.Unmarshal(m, cmd)
		fmt.Printf("%s\n%v\n", m, cmd)
		if err != nil{
			s.Send(err2byte(err))
			continue
		}
		c.Send(cmd)
	}
}

func err2byte(err error) []byte {
	return []byte(fmt.Sprintf(`{"error": "%s"}`, err))
}
