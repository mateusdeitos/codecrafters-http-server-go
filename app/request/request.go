package request

import (
	"errors"
	"strings"
)

type Request struct {
	Scheme  string
	Host    string
	Path    string
	Body    string
	Headers map[string]string
	Query   map[string]string
	Method  string
	Version string
}

type reader interface {
	Read(p []byte) (n int, err error)
}

func BuildRequest(c reader) (*Request, error) {
	s, err := getRequestStr(c)
	if err != nil {
		return nil, err
	}

	req := &Request{}
	if err := req.parseAddress(s); err != nil {
		return nil, err
	}

	if err := req.parseHeaders(s); err != nil {
		return nil, err
	}

	if err := req.parseBody(s); err != nil {
		return nil, err
	}

	return req, nil
}

func getRequestStr(c reader) (string, error) {
	buf := make([]byte, 1024)
	n, err := c.Read(buf)
	if err != nil {
		return "", err
	}

	return string(buf[:n]), nil
}

func (r *Request) parseAddress(rs string) error {
	a := strings.Split(rs, "\r\n")
	rl := strings.Split(a[0], " ")

	if len(rl) != 3 {
		return errors.New("invalid request, missing method, path and version")
	}

	r.Method = rl[0]
	r.Version = rl[2]

	path := rl[1]
	pos := strings.Index(path, "?")
	var queryStr string
	if pos == -1 {
		pos = len(path)
		queryStr = ""
	} else {
		queryStr = path[pos+1:]
	}

	pathStr := path[:pos]
	r.Path = pathStr

	r.Query = make(map[string]string)

	if queryStr != "" {
		for _, v := range strings.Split(queryStr, "&") {
			kv := strings.Split(v, "=")
			if len(kv) == 2 {
				r.Query[kv[0]] = kv[1]
			}
		}
	}

	return nil
}

func (r *Request) parseHeaders(rs string) error {
	a := strings.Split(rs, "\r\n")
	r.Headers = make(map[string]string)
	for _, h := range a {
		if h == "" {
			continue
		}

		kv := strings.Split(h, ":")
		if len(kv) != 2 {
			continue
		}

		k := strings.Trim(kv[0], " ")
		v := strings.Trim(kv[1], " ")

		switch k {
		case "Host":
			r.Host = v
		default:
			r.Headers[k] = v
		}
	}

	return nil
}

func (r *Request) parseBody(rs string) error {
	a := strings.Split(rs, "\r\n\r\n")

	if len(a) > 1 {
		r.Body = a[1]
	} else {
		r.Body = ""
	}

	return nil
}
