package http

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/skycoin/bbs/src/misc/inform"
)

const (
	logPrefix     = "HTTPSERVER"
	localhost     = "27.0.0.1"
	indexFileName = "index.html"
)

// ServerConfig represents a HTTP server configuration file.
type ServerConfig struct {
	Port      *int
	StaticDir *string
	EnableGUI *bool
}

// Server represents an HTTP Server that serves static files and API
type Server struct {
	c    *ServerConfig
	l    *log.Logger
	net  net.Listener
	mux  *http.ServeMux
	api  *Gateway
	quit chan struct{}
}

//NewServer creates a new server
func NewServer(config *ServerConfig, api *Gateway) (*Server, error) {
	server := &Server{
		c:    config,
		l:    inform.NewLogger(true, os.Stdout, logPrefix),
		mux:  http.NewServeMux(),
		api:  api,
		quit: make(chan struct{}),
	}

	var e error
	if *config.StaticDir, e = filepath.Abs(*config.StaticDir); e != nil {
		return nil, e
	}

	host := fmt.Sprintf("%s:%d", localhost, *config.Port)
	if server.net, e = net.Listen("tcp", host); e != nil {
		return nil, e
	}
	if e := server.prepareMux(); e != nil {
		return nil, e
	}
	go server.serve()
	return server, e
}

func (s *Server) prepareMux() error {
	if *s.c.EnableGUI {
		if e := s.prepareStatic(); e != nil {
			return e
		}
	}
	return s.api.host(s.mux)
}

func (s *Server) prepareStatic() error {
	appLoc := *s.c.StaticDir

}
