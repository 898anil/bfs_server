package bfs_server

import (
	"bufio"
	"fmt"
	"net"
	"regexp"
	"strings"
)

type Request struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    []byte
}

type Response struct {
	Headers map[string]string
	Body    []byte
}

type Handler func(req *Request) *Response

type Server struct {
	router Router
	Port   string
}

type Route struct {
	Path    string
	Handler Handler
}

type Router struct {
	routes []Route
}

func NewRouter() *Router {
	return &Router{routes: make([]Route, 0)}
}

func (r *Router) Add(route Route) {
	r.routes = append(r.routes, route)
}

func (r *Router) Handle(req *Request) *Response {
	for _, route := range r.routes {
		if route.Path == req.Path || matchRegex(route.Path, req.Path) {
			return route.Handler(req)
		}
	}
	return &Response{
		Headers: map[string]string{
			"Status": "HTTP/1.1 404 Not Found",
		},
		Body: []byte("404 Not Found"),
	}
}

func matchRegex(pattern, path string) bool {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return regex.MatchString(path)
}

func (s *Server) ListenAndServe() {
	listener, err := net.Listen("tcp", ":"+s.Port)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	var buf [512]byte
	n, err := reader.Read(buf[:])
	if err != nil {
		panic(err)
	}
	fmt.Println(string(buf[:n]))
	request := parseRequest(string(buf[:n]))
	response := s.router.Handle(request)
	conn.Write([]byte(response.Headers["Status"] + "\r\n"))
	for header, value := range response.Headers {
		if header != "Status" {
			conn.Write([]byte(header + ": " + value + "\r\n"))
		}
	}
	conn.Write([]byte("\r\n"))
	conn.Write(response.Body)
}

func parseRequest(request string) *Request {
	lines := strings.Split(request, "\r\n")
	if len(lines) == 0 {
		return &Request{Method: "", Path: "", Headers: make(map[string]string), Body: []byte{}}
	}

	parts := strings.Split(lines[0], " ")
	method := ""
	path := ""
	if len(parts) > 0 {
		method = parts[0]
	}
	if len(parts) > 1 {
		path = parts[1]
	}

	headers := make(map[string]string)
	var bodyStart int
	for i := 1; i < len(lines); i++ {
		if lines[i] == "" {
			bodyStart = i + 1
			break
		}
		parts := strings.SplitN(lines[i], ": ", 2)
		if len(parts) == 2 {
			headers[parts[0]] = parts[1]
		}
	}

	body := []byte(strings.Join(lines[bodyStart:], "\r\n"))
	return &Request{
		Method:  method,
		Path:    path,
		Headers: headers,
		Body:    body,
	}
}
