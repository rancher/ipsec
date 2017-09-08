package server

import (
	"fmt"
	"net/http"

	"github.com/leodotcloud/log"

	"github.com/rancher/ipsec/backend"
)

// Server structure is used to the store backend information
type Server struct {
	Backend backend.Backend
}

// ListenAndServe is used to setup ping and reload handlers and
// start listening on the specified port
func (s *Server) ListenAndServe(listen string) error {
	http.HandleFunc("/ping", s.ping)
	http.HandleFunc("/v1/reload", s.reload)
	http.HandleFunc("/v1/loglevel", s.loglevel)
	log.Infof("Listening on %s", listen)
	err := http.ListenAndServe(listen, nil)
	if err != nil {
		log.Errorf("got error while ListenAndServe: %v", err)
	}
	return err
}

func (s *Server) ping(rw http.ResponseWriter, req *http.Request) {
	log.Debugf("Received ping request")
	rw.Write([]byte("OK"))
}

func (s *Server) reload(rw http.ResponseWriter, req *http.Request) {
	log.Debugf("Received reload request")
	msg := "Reloaded Configuration\n"
	if err := s.Backend.Reload(); err != nil {
		rw.WriteHeader(500)
		msg = fmt.Sprintf("Failed to reload configuration: %v\n", err)
	}

	rw.Write([]byte(msg))
}

func (s *Server) loglevel(rw http.ResponseWriter, req *http.Request) {
	// curl -X POST -d "level=debug" localhost:8111/v1/loglevel
	log.Debugf("Received loglevel request")
	if req.Method == http.MethodGet {
		level := log.GetLevel().String()
		rw.Write([]byte(fmt.Sprintf("%s\n", level)))
	}

	if req.Method == http.MethodPost {
		if err := req.ParseForm(); err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(fmt.Sprintf("Failed to parse form: %v\n", err)))
		}
		err := log.SetLevelString(req.Form.Get("level"))
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(fmt.Sprintf("Failed to set loglevel: %v\n", err)))
		} else {
			rw.Write([]byte("OK\n"))
		}
	}
}
