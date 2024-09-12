package main

import (
	"fmt"
	"net"
	"os"
	"regexp"

	"github.com/codecrafters-io/http-server-starter-go/app/request"
	"github.com/codecrafters-io/http-server-starter-go/app/response"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	for {
		go runListener(l)
	}

}

func runListener(l net.Listener) {

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		return
	}

	req, err := request.BuildRequest(conn)
	var resp *response.Response

	if err != nil {
		handleConnection(conn, response.New(400, err.Error()))
		return
	}

	if resp = indexRoute(req); resp != nil {
		handleConnection(conn, resp)
		return
	}

	if resp = echoRoute(req); resp != nil {
		handleConnection(conn, resp)
		return
	}

	if resp = userAgentRoute(req); resp != nil {
		handleConnection(conn, resp)
		return
	}

	handleConnection(conn, response.New(404, ""))
}

func handleConnection(conn net.Conn, resp *response.Response) {
	conn.Write([]byte(resp.String()))
	conn.Close()
}

func indexRoute(req *request.Request) *response.Response {
	if req.Method != "GET" {
		return nil
	}

	if req.Path != "/" {
		return nil
	}

	return response.New(200, "")
}

func echoRoute(req *request.Request) *response.Response {
	rx := regexp.MustCompile("^/echo/([^/]+)$")
	matches := rx.FindStringSubmatch(req.Path)
	if matches == nil {
		return nil
	}

	param := matches[1]
	if param == "" {
		return nil
	}

	return response.New(200, string(param))
}

func userAgentRoute(req *request.Request) *response.Response {
	if req.Method != "GET" {
		return nil
	}

	if req.Path != "/user-agent" {
		return nil
	}

	return response.New(200, req.Headers["User-Agent"])
}
