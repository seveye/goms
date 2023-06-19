package main

import (
	"fmt"

	"gitee.com/jkkkls/goms/util/bytes_cache"
)

func main() {
	x := bytes_cache.Get(4, 10)
	fmt.Println(len(x), cap(x), &x[0])

	bytes_cache.Put(x)

	y := bytes_cache.Get(0, 1)
	y = append(y, 1)
	fmt.Println(len(y), cap(y), &y[0])

	y1 := bytes_cache.Get(9)
	fmt.Println(len(y1), cap(y1), &y1[0])
}
