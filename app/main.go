package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
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

	rootDir := "tmp"
	if len(os.Args) > 2 {
		rootDir = os.Args[2]
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			return
		}

		go runListener(conn, rootDir)
	}

}

func runListener(conn net.Conn, rootDir string) {
	req, err := request.BuildRequest(conn)
	var resp *response.Response

	if err != nil {
		handleConnection(conn, req, response.New(400, []byte(err.Error())))
		return
	}

	if resp = indexRoute(req); resp != nil {
		handleConnection(conn, req, resp)
		return
	}

	if resp = echoRoute(req); resp != nil {
		handleConnection(conn, req, resp)
		return
	}

	if resp = userAgentRoute(req); resp != nil {
		handleConnection(conn, req, resp)
		return
	}

	if resp = fileRoute(rootDir, req); resp != nil {
		handleConnection(conn, req, resp)
		return
	}

	if resp = createFileRoute(rootDir, req); resp != nil {
		handleConnection(conn, req, resp)
		return
	}

	handleConnection(conn, req, response.New(404, nil))
}

func handleConnection(conn net.Conn, req *request.Request, resp *response.Response) {
	resp.Compress(req.Headers["Accept-Encoding"])
	resp.Write(conn)
	conn.Close()
}

func indexRoute(req *request.Request) *response.Response {
	if req.Method != "GET" {
		return nil
	}

	if req.Path != "/" {
		return nil
	}

	return response.New(200, nil)
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

	return response.New(200, []byte(param))
}

func userAgentRoute(req *request.Request) *response.Response {
	if req.Method != "GET" {
		return nil
	}

	if req.Path != "/user-agent" {
		return nil
	}

	return response.New(200, []byte(req.Headers["User-Agent"]))
}

func fileRoute(rootDir string, req *request.Request) *response.Response {
	if req.Method != "GET" {
		return nil
	}

	rx := regexp.MustCompile("^/files/([^/]+)$")
	matches := rx.FindStringSubmatch(req.Path)
	if matches == nil {
		return nil
	}

	filename := matches[1]
	if filename == "" {
		return nil
	}

	filename = filepath.Join(rootDir, filename)

	s, err := os.Stat(filename)
	if err != nil && os.IsNotExist(err) {
		return response.New(404, nil)
	}

	if s.IsDir() {
		return response.New(404, nil)
	}

	contents, _ := os.ReadFile(filename)

	r := response.New(200, contents)
	r.AddHeader("Content-Type", "application/octet-stream")
	r.AddHeader("Content-Length", fmt.Sprintf("%d", len(contents)))
	return r
}

func createFileRoute(rootDir string, req *request.Request) *response.Response {
	if req.Method != "POST" {
		return nil
	}

	rx := regexp.MustCompile("^/files/([^/]+)$")
	matches := rx.FindStringSubmatch(req.Path)
	if matches == nil {
		return nil
	}

	filename := matches[1]
	if filename == "" {
		return nil
	}

	_, err := os.Stat(rootDir)
	if err != nil && os.IsNotExist(err) {
		err = os.Mkdir(rootDir, 0644)
		if err != nil {
			return response.New(400, []byte(err.Error()))
		}
	}

	filename = filepath.Join(rootDir, filename)

	s, err := os.Stat(filename)
	if err != nil && os.IsExist(err) {
		return response.New(400, nil)
	}

	if s != nil && s.IsDir() {
		return response.New(400, nil)
	}

	err = os.WriteFile(filename, []byte(req.Body), 0644)
	if err != nil {
		return response.New(400, []byte(err.Error()))
	}

	r := response.New(201, nil)
	return r
}
