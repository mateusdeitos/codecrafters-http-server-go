package response

import (
	"fmt"
)

type status int

func (s status) String() string {
	switch s {
	case 200:
		return "200 OK"
	case 400:
		return "400 Bad Request"
	case 404:
		return "404 Not Found"
	default:
		return "500 Internal Server Error"
	}
}

type Response struct {
	Headers map[string]string
	Status  status
	Body    string
}

func New(s int, body string) *Response {
	r := &Response{
		Headers: make(map[string]string),
		Status:  status(s),
		Body:    body,
	}

	r.AddHeader("Content-Length", fmt.Sprintf("%d", len(r.Body)))
	r.AddHeader("Content-Type", "text/plain")

	return r
}

func (s *Response) AddHeader(key, value string) {
	s.Headers[key] = value
}

func (r *Response) String() string {
	resp := fmt.Sprintf("HTTP/1.1 %s\r\n", r.Status.String())
	for k, v := range r.Headers {
		resp += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	resp += "\r\n"
	resp += r.Body

	return resp
}
