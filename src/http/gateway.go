package http

import (
	"log"
	"net/http"
	"os"

	"spaco.io/cosmos/src/misc/inform"
)

// Gateway represents what is exposed to HTTP interface.
type Gateway struct {
	l        *log.Logger
	QuitChan chan int
}

func (g *Gateway) host(mux *http.ServeMux) error {
	g.l = inform.NewLogger(true, os.Stdout, "")
	return nil
}
