package main

import (
	"github.com/seveye/goms/watch"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	go http.ListenAndServe("0.0.0.0:7070", nil)

	server := watch.NewWatchServer()
	server.Start(":12345")
}
