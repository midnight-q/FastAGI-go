package AGIServer

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type Server struct {
	Addr         string
	IdleTimeout  time.Duration
	MaxReadBytes int64

	listener   net.Listener
	conns      map[*conn]struct{}
	mu         sync.Mutex
	inShutdown bool
	handlers   map[string]Route
}

type Route func(Request)

func (srv *Server) AddRoute(name string, route Route) error {
	if srv.handlers == nil {
		srv.handlers = make(map[string]Route)
	}

	_, isExist := srv.handlers[name]
	if isExist {
		return fmt.Errorf("this route alredy exist")
	}
	srv.handlers[name] = route
	return nil
}

func (srv *Server) ListenAndServe() {
	addr := srv.Addr
	if addr == "" {
		addr = ":8080"
	}
	log.Printf("starting server on %v\n", addr)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}
	defer listener.Close()
	srv.listener = listener
	for {
		// should be guarded by mu
		if srv.inShutdown {
			break
		}
		newConn, err := listener.Accept()
		if err != nil {
			log.Printf("error accepting connection %v", err)
			continue
		}
		//log.Printf("accepted connection from %v", newConn.RemoteAddr())
		conn := &conn{
			Conn:          newConn,
			IdleTimeout:   srv.IdleTimeout,
			MaxReadBuffer: srv.MaxReadBytes,
		}
		srv.trackConn(conn)
		_ = conn.SetDeadline(time.Now().Add(conn.IdleTimeout))
		go srv.handle(conn)
	}
	return
}

func (srv *Server) trackConn(c *conn) {
	defer srv.mu.Unlock()
	srv.mu.Lock()
	if srv.conns == nil {
		srv.conns = make(map[*conn]struct{})
	}
	srv.conns[c] = struct{}{}
}

func (srv *Server) handle(conn *conn) {
	defer func() {
		//log.Printf("closing connection from %v", conn.RemoteAddr())
		_ = conn.Close()
		srv.deleteConn(conn)
	}()
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	scanner := bufio.NewScanner(r)
	var str strings.Builder

	sc := make(chan bool)
	for {
		go func(s chan bool) {
			s <- scanner.Scan()
		}(sc)
		select {
		case scanned := <-sc:
			if !scanned {
				if err := scanner.Err(); err != nil {
					return
				}
				return
			}
			if scanner.Text() == "" {
				data := ParseText(str.String())
				request := Request{data, w, r}
				route, isExist := data["agi_network_script"]
				if !isExist {
					return
				}
				handler, isExist := srv.handlers[route]
				if !isExist {
					return
				}
				handler(request)
				_ = conn.Close()
			} else {
				str.WriteString("\n")
				str.WriteString(scanner.Text())
			}
		}
	}
}

func (srv *Server) deleteConn(conn *conn) {
	defer srv.mu.Unlock()
	srv.mu.Lock()
	delete(srv.conns, conn)
}

func (srv *Server) Shutdown() {
	// should be guarded by mu
	srv.inShutdown = true
	log.Println("shutting down...")
	_ = srv.listener.Close()
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			log.Printf("waiting on %v connections", len(srv.conns))
		}
		if len(srv.conns) == 0 {
			return
		}
	}
}

func ParseText(text string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		data := strings.Split(line, ": ")
		if len(data) < 2 {
			continue
		}
		result[data[0]] = data[1]
	}
	return result
}
