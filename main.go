package main

// reverseproxy implements HTTPS termination and reverse proxying to localhost ports.
// It reads /etc/reverseproxy/hosts.json to determine which hosts to proxy to which ports.
// It listens to requests on port 80 and 443.

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/mikerybka/reverseproxy/pkg/web"
	"github.com/mikerybka/util"
)

const workdir = "/etc/reverseproxy"

func main() {
	logdir := filepath.Join(workdir, "logs")
	err := os.MkdirAll(logdir, os.ModePerm)
	if err != nil {
		panic(err)
	}
	certdir := filepath.Join(workdir, "certs")
	err = os.MkdirAll(certdir, os.ModePerm)
	if err != nil {
		panic(err)
	}
	hostsfile := filepath.Join(workdir, "hosts.json")
	email := os.Getenv("EMAIL")
	h := &handler{
		hostsfile: hostsfile,
		logdir:    logdir,
	}
	err = util.ServeHTTPS(h, email, certdir)
	if err != nil {
		panic(err)
	}
}

type handler struct {
	hostsfile string
	logdir    string
}

func hosts() map[string]string {
	path := "/etc/reverseproxy/hosts.json"
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var hosts map[string]string
	err = json.Unmarshal(data, &hosts)
	if err != nil {
		panic(err)
	}
	return hosts
}

// ServeHTTP implements http.Handler.
// It simply proxies requests to the localhost port specified in the hosts file.
// The hosts file will not be read more than once every 5 seconds.
// If the host is not found in the hosts file, it returns a 404.
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logRequest(r, h.logdir)
	backendPort, ok := hosts()[r.Host]
	if !ok {
		http.NotFound(w, r)
		return
	}
	proxy := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Scheme = "http"
			r.URL.Host = "localhost:" + backendPort
		},
	}
	proxy.ServeHTTP(w, r)
}

func logRequest(r *http.Request, logdir string) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	l := web.Request{
		IP:      r.RemoteAddr,
		Method:  r.Method,
		Host:    r.Host,
		Path:    r.URL.Path,
		Query:   r.URL.Query(),
		Headers: r.Header,
		Body:    body,
	}
	b, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	timestamp := time.Now().UnixNano()
	logFile := filepath.Join(logdir, strconv.Itoa(int(timestamp)))
	return os.WriteFile(logFile, b, os.ModePerm)
}
