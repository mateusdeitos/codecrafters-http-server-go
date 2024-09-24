package request

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ParseConn(t *testing.T) {

	cases := []struct {
		conn    reader
		method  string
		host    string
		path    string
		query   map[string]string
		version string
		headers map[string]string
		body    string
	}{
		{
			conn: &fakeConn{
				createRequestString([]string{
					"GET /test?key=value HTTP/1.1",
					"Host: localhost",
					"User-Agent: curl/7.68.0",
					"Accept: */*",
				}),
			},
			method: "GET",
			host:   "localhost",
			path:   "/test",
			query: map[string]string{
				"key": "value",
			},
			version: "HTTP/1.1",
			headers: map[string]string{
				"User-Agent": "curl/7.68.0",
				"Accept":     "*/*",
			},
			body: "",
		},
		{
			conn: &fakeConn{
				createRequestString([]string{
					"POST /todo HTTP/1.1",
					"Host: google.com",
					"Accept: */*",
					"User-Agent: curl/7.68.0",
					"Content-Type: application/json",
					"Content-Length: 5",
				}),
			},
			method:  "POST",
			host:    "google.com",
			path:    "/todo",
			query:   map[string]string{},
			version: "HTTP/1.1",
			headers: map[string]string{
				"User-Agent":     "curl/7.68.0",
				"Accept":         "*/*",
				"Content-Type":   "application/json",
				"Content-Length": "5",
			},
			body: "",
		},
		{
			conn: &fakeConn{
				createRequestString([]string{
					"POST /files/apple_apple_grape_grape HTTP/1.1",
					"Host: localhost",
					"Content-Length: 55",
					"Content-Type: application/octet-stream",
					"",
					"blueberry grape apple mango grape mango pineapple grape",
				}),
			},
			method:  "POST",
			host:    "localhost",
			path:    "/files/apple_apple_grape_grape",
			query:   map[string]string{},
			version: "HTTP/1.1",
			headers: map[string]string{
				"Content-Type":   "application/octet-stream",
				"Content-Length": "55",
			},
			body: "blueberry grape apple mango grape mango pineapple grape",
		},
	}

	for i, c := range cases {
		req, err := BuildRequest(c.conn)
		if err != nil {
			t.Errorf("Unexpected error: %s, case %d", err, i)
			return
		}

		assert.Equal(t, c.method, req.Method, fmt.Sprintf("case %d", i))
		assert.Equal(t, c.host, req.Host, fmt.Sprintf("case %d", i))
		assert.Equal(t, c.path, req.Path, fmt.Sprintf("case %d", i))
		assert.Equal(t, c.query, req.Query, fmt.Sprintf("case %d", i))
		assert.Equal(t, c.version, req.Version, fmt.Sprintf("case %d", i))
		assert.Equal(t, c.headers, req.Headers, fmt.Sprintf("case %d", i))
		assert.Equal(t, c.body, req.Body, fmt.Sprintf("case %d", i))
	}

}

type fakeConn struct {
	requestString string
}

func (c *fakeConn) Read(p []byte) (n int, err error) {
	n = copy(p, []byte(c.requestString))
	return n, nil
}

func createRequestString(parts []string) string {
	return strings.Join(parts, "\r\n")
}
