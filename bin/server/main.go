package main

import (
	"github.com/domac/sckv"
	"log"
)

func main() {

	server := sckv.NewServer("0.0.0.0:6380", sckv.HandlerFunc(serverSideHandler), 10)
	server.ListenAndServe()
}

func serverSideHandler(session *sckv.Session) {
	for {
		cmds, err := session.Receive()
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("cmds = %s\n", cmds)
		session.WriteOK()
	}
}
