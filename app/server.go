package main

import (
	"fmt"
	"net"
	"os"
	"regexp"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	//
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading: ", err.Error())
		os.Exit(1)
	}
	fmt.Println(string(buf[:n]))

	// extract the request target
	r := regexp.MustCompile("[GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS] (.*) HTTP/1.1\r\n")
	matches := r.FindAllStringSubmatch(string(buf[:n]), 1)
	if len(matches) == 0 {
		conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
		conn.Close()
		return
	}

	path := matches[0][1]
	fmt.Printf("path: %s\n", path)
	if path != "/" {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		conn.Close()
		return
	}

	conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	conn.Close()
}
