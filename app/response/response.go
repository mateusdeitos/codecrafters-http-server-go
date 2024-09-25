package response

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"
)

type status int

const (
	gzipEncoding = "gzip"
)

func (s status) String() string {
	switch s {
	case 200:
		return "200 OK"
	case 201:
		return "201 Created"
	case 204:
		return "204 No Content"
	case 400:
		return "400 Bad Request"
	case 404:
		return "404 Not Found"
	default:
		panic("invalid status code")
	}
}

type Response struct {
	Headers map[string]string
	Status  status
	Body    []byte
}

func New(s int, body []byte) *Response {
	r := &Response{
		Headers: make(map[string]string),
		Status:  status(s),
		Body:    body,
	}

	bl := len(r.Body)
	r.AddHeader("Content-Length", fmt.Sprintf("%d", bl))
	if bl > 0 {
		r.AddHeader("Content-Type", "text/plain")
	}

	return r
}

func (s *Response) AddHeader(key, value string) {
	s.Headers[key] = value
}

func (r *Response) Compress(acceptEncoding string) {
	type compressHandler func(string) ([]byte, error)

	supportedEncodings := map[string]compressHandler{
		gzipEncoding: func(s string) ([]byte, error) {
			return gzipCompress(r.Body)
		},
	}

	for _, v := range strings.Split(acceptEncoding, ",") {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}

		h, ok := supportedEncodings[v]
		if ok {
			b, err := h(v)
			if err != nil {
				continue
			}

			r.Body = b
			r.Headers["Content-Encoding"] = v
			r.Headers["Content-Length"] = fmt.Sprintf("%d", len(r.Body))
			break
		}
	}
}

func (r *Response) Write(w io.Writer) error {
	// Write the response
	resp := fmt.Sprintf("HTTP/1.1 %s\r\n", r.Status.String())
	for k, v := range r.Headers {
		resp += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	resp += "\r\n"

	// Write the response headers
	if _, err := w.Write([]byte(resp)); err != nil {
		return err
	}

	if r.Body != nil {
		_, err := w.Write(r.Body)
		return err
	}

	return nil
}

func gzipCompress(body []byte) ([]byte, error) {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	if _, err := w.Write(body); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	// convert buffer output to hexadecimal representation
	output := b.Bytes()

	return output, nil
}
