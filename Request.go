package AGIServer

import (
	"bufio"
	"fmt"
	"strings"
)

type Request struct {
	Params map[string]string
	writer *bufio.Writer
	reader *bufio.Reader
}

func (r *Request) SendCommand(command string) {
	_, err := r.writer.WriteString(command + "\n")
	if err != nil {
		fmt.Println(err)
	}
	_ = r.writer.Flush()

	for {
		line, _, _ := r.reader.ReadLine()
		data := strings.Split(string(line), " ")
		if data[0] != "100" {
			break
		}
	}
}

func (r *Request) Busy() {
	r.SendCommand("EXEC Hangup 17")
}
