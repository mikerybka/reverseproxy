package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/library-development/go-web"
	"golang.org/x/crypto/acme/autocert"
)

func main() {
	flag.Parse()
	configfile := "/etc/reverseproxy/config.json"
	certdir := "/etc/ssl/certs"
	logdir := "/var/log/reverseproxy"
	email := "merybka@gmail.com"
	h := &handler{
		configfile: configfile,
		logdir:     logdir,
	}
	manager := autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache(certdir),
		HostPolicy: func(_ context.Context, host string) error {
			h.readConfig()
			_, ok := h.config[host]
			if ok {
				return nil
			}
			return fmt.Errorf("host %q not allowed", host)
		},
		Email: email,
	}
	l := manager.Listener()
	err := http.Serve(l, h)
	panic(err)
}

type handler struct {
	configfile     string
	logdir         string
	config         map[string]string
	configLastRead time.Time
}

// readConfig reads the config file and stores it in h.config.
// If the config file has been read in the last 5 seconds, it
// does not read it again.
func (h *handler) readConfig() error {
	now := time.Now()
	if now.Sub(h.configLastRead) < 5*time.Second {
		return nil
	}
	data, err := os.ReadFile(h.configfile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &h.config)
	if err != nil {
		return err
	}
	return nil
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	web.LogRequest(r, h.logdir)
	err := h.readConfig()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	backendPort, ok := h.config[r.Host]
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
