package http

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type HttpServer struct {
	server  *http.Server
	port    int
	recipes map[string]Recipe
}

// NewHttpServer creates and returns a new HttpServer with a listener on the specified port
func NewHttpServer(port int) (HttpServer, error) {
	s := HttpServer{server: &http.Server{Addr: fmt.Sprintf(":%d", port)},
		port:    port,
		recipes: make(map[string]Recipe),
	}

	mux := http.NewServeMux()
	s.server.Handler = mux

	return s, nil
}

// Start initiates the HTTP listener of the received HttpServer
func (s *HttpServer) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		fmt.Printf("starting HTTP listener on port %d\n", s.port)
		defer fmt.Printf("HTP listener on port %d stopped\n", s.port)
		if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Printf("http server error: %v\n", err)
		}
	}()
}

// Stop terminates the HTTP listener of the received HttpServer
func (s *HttpServer) Stop() {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := s.server.Shutdown(ctxTimeout); err != nil {
		panic(err) // failure/timeout shutting down the server gracefully
	}
}
