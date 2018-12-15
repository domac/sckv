package main

import (
	"github.com/domac/sckv"
	_ "github.com/domac/sckv/store/stdmap"
	"log"
)

var Cache sckv.SCCache

func main() {
	var err error
	Cache, err = sckv.New("MapCache")
	if err != nil {
		log.Println(err)
		return
	}
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
		if "get" == string(cmds[0].Args[0]) {
			session.WriteValue(Cache.Get(cmds[0].Args[1]))
		} else if "set" == string(cmds[0].Args[0]) {
			Cache.Set(cmds[0].Args[1], cmds[0].Args[2])
			session.WriteOK()
		} else {
			session.WriteValue([]byte("command error"))
		}
	}
}
