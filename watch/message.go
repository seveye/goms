// Copyright 2017 guangbo. All rights reserved.

//watch模块消息协议，实现参考redis协议
//使用示例参考gitee.com/goxiang2/server/example/watch
package watch

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

type WatchMessage struct {
	Cmd    string
	Seq    int
	Values []string
}

func ReadWatchMessage(r *bufio.Reader) (*WatchMessage, error) {
	message := &WatchMessage{}
	var (
		nStr, value string
		err         error
		n           int
	)
	message.Cmd, err = r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	message.Cmd = strings.TrimSpace(message.Cmd)

	nStr, err = r.ReadString('\n')
	nStr = strings.Trim(nStr, "\n")
	if err != nil {
		return nil, err
	}
	message.Seq, _ = strconv.Atoi(nStr)

	nStr, err = r.ReadString('\n')
	nStr = strings.Trim(nStr, "\n")
	if err != nil {
		return nil, err
	}
	n, _ = strconv.Atoi(nStr)

	for i := 0; i < n; i++ {
		value, err = r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		value = strings.Trim(value, "\n")
		message.Values = append(message.Values, value)
	}
	// log.Println("ReadWatchMessage", message.Cmd, message.Seq, message.Values)
	return message, nil
}

func WriteWatchMessage(w *bufio.Writer, message *WatchMessage) error {
	// log.Println("WriteWatchMessage", message.Cmd, message.Seq, message.Values)
	_, err := w.WriteString(fmt.Sprintf("%v\n", message.Cmd))
	if err != nil {
		return err
	}
	_, err = w.WriteString(fmt.Sprintf("%v\n", message.Seq))
	if err != nil {
		return err
	}
	n := len(message.Values)
	_, err = w.WriteString(fmt.Sprintf("%v\n", n))
	if err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		_, err = w.WriteString(fmt.Sprintf("%v\n", message.Values[i]))
		if err != nil {
			return err
		}
	}
	w.Flush()
	return nil
}
