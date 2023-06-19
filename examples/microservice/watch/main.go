// Copyright 2017 guangbo. All rights reserved.
package main

import (
	"fmt"

	"github.com/seveye/goms/watch"
)

func main() {
	server := watch.NewWatchServer()
	err := server.Start(":12345")
	fmt.Println(err)
}
