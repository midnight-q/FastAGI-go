package AGIServer

import "bufio"

type Request struct {
	Params map[string]string
	writer *bufio.Writer
}

func (r *Request) SendCommand(rawCommand string) {
	command := rawCommand + "\n"
	r.writer.WriteString(command)
	r.writer.Flush()
}
