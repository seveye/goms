package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

	"gitee.com/jkkkls/goms/watch"
	// "time"
)

func main() {
	go http.ListenAndServe("0.0.0.0:6060", nil)

	client, err := watch.NewWatchClient("127.0.0.1:19876")
	if err != nil {
		log.Println(err)
		return
	}
	go func() {
		for {
			key, field, value := client.Watch()
			log.Println("watch", key, field, value)
		}
	}()

	client.Start()

	// for i := 0; i <= 100; i++ {
	// 	time.Sleep(2 * time.Second)
	// 	key := fmt.Sprintf("key%v", i%3)
	// 	field := fmt.Sprintf("field%v", i)
	// 	value := fmt.Sprintf("value%v", i)
	// 	client.Hset(key, field, value)

	// 	//log.Println("hget", key, field, client.Hget(key, field))
	// }

	info := client.Hget("key1", "field1")
	fmt.Println(info)

	client.Shutdown()
}
