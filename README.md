# FastAGI-go
Asterisk FastAGI server for GoLang

### Install

```
go get https://github.com/BlayD91/FastAGI-go
```

### Usage

```go
import "time"

func callHandler(r Request) {
	r.SendCommand("ANSWER")

	if r.Params["agi_language"] == "en" {
		r.SendCommand("SET LANG_CODE 1")
	} else {
		r.SendCommand("SET LANG_CODE 0")
	}

	r.SendCommand("HANGUP")
}

func main() {
	server := NewServer(":8000", 10*time.Second, 10*1024)
	err := server.AddRoute("inner_call", callHandler)
	if err != nil {
		panic(err)
	}
	go server.ListenAndServe()

	time.Sleep(10 * time.Second)
}
```
