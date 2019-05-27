package AGIServer

import "bufio"

type Request struct {
	Params map[string]string
	writer *bufio.Writer
}

func (r *Request) SendCommand(command string) {
	_, _ = r.writer.WriteString(command)
	_ = r.writer.Flush()
}
