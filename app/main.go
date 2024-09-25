package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"syscall"

	"github.com/codecrafters-io/http-server-starter-go/app/request"
	"github.com/codecrafters-io/http-server-starter-go/app/response"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit
var echoRouteRegex = regexp.MustCompile("^/echo/([^/]+)$")
var filesRouteRegex = regexp.MustCompile("^/files/([^/]+)$")

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	rootDir := "tmp"
	if len(os.Args) > 2 {
		rootDir = os.Args[2]
	}

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				select {
				case <-stop:
					return
				default:
					fmt.Println("Error accepting connection: ", err.Error())
					return
				}
			}

			go runListener(conn, rootDir)
		}

	}()

	<-stop
	fmt.Println("Shutting down...")

}

func runListener(conn net.Conn, rootDir string) {
	req, err := request.BuildRequest(conn)
	var resp *response.Response

	if err != nil {
		processRequest(conn, req, response.New(http.StatusBadRequest, []byte(err.Error())))
		return
	}

	if resp = indexRoute(req); resp != nil {
		processRequest(conn, req, resp)
		return
	}

	if resp = echoRoute(req); resp != nil {
		processRequest(conn, req, resp)
		return
	}

	if resp = userAgentRoute(req); resp != nil {
		processRequest(conn, req, resp)
		return
	}

	if resp = fileRoute(rootDir, req); resp != nil {
		processRequest(conn, req, resp)
		return
	}

	if resp = createFileRoute(rootDir, req); resp != nil {
		processRequest(conn, req, resp)
		return
	}

	processRequest(conn, req, response.New(http.StatusNotFound, nil))
}

func processRequest(conn net.Conn, req *request.Request, resp *response.Response) {
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

	return response.New(http.StatusOK, nil)
}

func echoRoute(req *request.Request) *response.Response {
	matches := echoRouteRegex.FindStringSubmatch(req.Path)
	if matches == nil {
		return nil
	}

	param := matches[1]
	if param == "" {
		return nil
	}

	return response.New(http.StatusOK, []byte(param))
}

func userAgentRoute(req *request.Request) *response.Response {
	if req.Method != "GET" {
		return nil
	}

	if req.Path != "/user-agent" {
		return nil
	}

	return response.New(http.StatusOK, []byte(req.Headers["User-Agent"]))
}

func fileRoute(rootDir string, req *request.Request) *response.Response {
	if req.Method != "GET" {
		return nil
	}

	matches := filesRouteRegex.FindStringSubmatch(req.Path)
	if matches == nil {
		return nil
	}

	filename := matches[1]
	if filename == "" {
		return nil
	}

	filename = filepath.Join(rootDir, filename)

	s, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return response.New(http.StatusNotFound, nil)
		}

		return response.New(http.StatusInternalServerError, []byte("Internal server error"))
	}

	if s.IsDir() {
		return response.New(http.StatusNotFound, nil)
	}

	contents, _ := os.ReadFile(filename)

	r := response.New(http.StatusOK, contents)
	r.AddHeader("Content-Type", "application/octet-stream")
	r.AddHeader("Content-Length", fmt.Sprintf("%d", len(contents)))
	return r
}

func createFileRoute(rootDir string, req *request.Request) *response.Response {
	if req.Method != "POST" {
		return nil
	}

	matches := filesRouteRegex.FindStringSubmatch(req.Path)
	if matches == nil {
		return nil
	}

	filename := matches[1]
	if filename == "" {
		return nil
	}

	err := os.MkdirAll(rootDir, 0644)
	if err != nil {
		return response.New(http.StatusBadRequest, []byte(err.Error()))
	}

	filename = filepath.Join(rootDir, filename)

	s, err := os.Stat(filename)
	if err != nil && os.IsExist(err) {
		return response.New(http.StatusConflict, nil)
	}

	if s != nil && s.IsDir() {
		return response.New(http.StatusBadRequest, nil)
	}

	err = os.WriteFile(filename, []byte(req.Body), 0644)
	if err != nil {
		return response.New(http.StatusBadRequest, []byte(err.Error()))
	}

	r := response.New(http.StatusCreated, nil)
	return r
}
