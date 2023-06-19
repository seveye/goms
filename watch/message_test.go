package watch

import (
	"bufio"
	"bytes"
	"testing"
)

func Test_Message(t *testing.T) {
	msg := &WatchMessage{
		Cmd: "init",
	}
	msg.Values = append(msg.Values, "aaaaa")
	msg.Values = append(msg.Values, "bbbbb")

	b := bytes.NewBuffer([]byte{})
	w := bufio.NewWriter(b)
	err := WriteWatchMessage(w, msg)

	r := bufio.NewReader(b)
	msg1, err := ReadWatchMessage(r)

	if err != nil || msg1.Cmd != msg.Cmd || len(msg.Values) != len(msg1.Values) || msg.Values[0] != msg1.Values[0] || msg.Values[1] != msg1.Values[1] {
		t.Error("Test_Message fail")
	}

}
