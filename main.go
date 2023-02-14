package main

// reverseproxy implements HTTPS termination and reverse proxying to localhost ports.
// It reads /etc/reverseproxy/hosts.json to determine which hosts to proxy to which ports.
// It listens to requests on port 80 and 443.

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

const workdir = "/etc/reverseproxy"
const configCooldown = 5 * time.Second

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
	manager := autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache(certdir),
		HostPolicy: func(_ context.Context, host string) error {
			h.readHosts()
			if err != nil {
				panic(err)
			}
			_, ok := h.hosts[host]
			if ok {
				return nil
			}
			return fmt.Errorf("host %q not allowed", host)
		},
		Email: email,
	}
	l := manager.Listener()
	err = http.Serve(l, h)
	panic(err)
}

type handler struct {
	hostsfile     string
	logdir        string
	hosts         map[string]string
	hostsLastRead time.Time
}

// readHosts reads the hosts file and stores it in h.hosts.
// If the hosts file has been read in the last 5 seconds, it
// does not read it again.
func (h *handler) readHosts() {
	now := time.Now()
	if now.Sub(h.hostsLastRead) < configCooldown {
		return
	}
	data, _ := os.ReadFile(h.hostsfile)
	json.Unmarshal(data, &h.hosts)
}

// ServeHTTP implements http.Handler.
// It simply proxies requests to the localhost port specified in the hosts file.
// The hosts file will not be read more than once every 5 seconds.
// If the host is not found in the hosts file, it returns a 404.
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logRequest(r, h.logdir)
	h.readHosts()
	backendPort, ok := h.hosts[r.Host]
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

type RequestLog struct {
	IP      string              `json:"ip"`
	Method  string              `json:"method"`
	Host    string              `json:"host"`
	Path    string              `json:"path"`
	Query   map[string][]string `json:"query"`
	Headers map[string][]string `json:"headers"`
	Body    []byte              `json:"body"`
}

func logRequest(r *http.Request, logdir string) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	l := RequestLog{
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
