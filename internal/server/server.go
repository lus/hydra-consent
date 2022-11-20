package server

import (
	"context"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type Server struct {
	httpServer *http.Server
	Address    string
}

func (server *Server) Run() error {
	router := chi.NewRouter()

	httpServer := &http.Server{
		Addr:    server.Address,
		Handler: router,
	}
	return httpServer.ListenAndServe()
}

func (server *Server) Stop(ctx context.Context) error {
	if server.httpServer == nil {
		return nil
	}
	if err := server.httpServer.Shutdown(ctx); err != nil {
		return err
	}
	server.httpServer = nil
	return nil
}
