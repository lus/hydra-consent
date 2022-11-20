package server

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	kratos "github.com/ory/client-go"
	hydra "github.com/ory/hydra-client-go"
	"net/http"
)

type Server struct {
	httpServer *http.Server
	Address    string
	Hydra      *hydra.APIClient
	Kratos     *kratos.APIClient
}

func (server *Server) Run() error {
	router := chi.NewRouter()
	router.Use(middleware.Recoverer)

	cnt := &controller{
		Hydra:  server.Hydra,
		Kratos: server.Kratos,
	}
	router.Get("/consent", cnt.Endpoint)

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
