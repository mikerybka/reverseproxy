package reverseproxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Server struct {
	Routes   map[string]string
	NotFound http.Handler
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()
	for k, v := range s.Routes {
		u, err := url.Parse(v)
		if err != nil {
			panic(err)
		}
		mux.Handle(k, httputil.NewSingleHostReverseProxy(u))
	}
	mux.Handle("/", s.NotFound)
	mux.ServeHTTP(w, r)
}
