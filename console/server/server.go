// Copyright (C) 2020 Creditor Corp. Group.
// See LICENSE for copying information.

package server

import (
	"context"
	"encoding/json"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zeebo/errs"
	"golang.org/x/sync/errgroup"

	"paxful/console"
	"paxful/internal/logger"
)

var (
	// Error is an error class for internal payment console http server error.
	Error = errs.Class("payment console web server error")
)

// Config contains configuration for paxful payment http server.
type Config struct {
	Address string `json:"address" help:"url paxful payments web server" default:"127.0.0.1:8081"`
}

// Server represents main admin portal http server with all endpoints.
//
// architecture: Endpoint
type Server struct {
	log    logger.Logger
	config Config

	service *console.Service

	server   http.Server
	listener net.Listener
}

// NewServer returns new instance of paxful trading console.
func NewServer(log logger.Logger, service *console.Service, config Config, listener net.Listener) (*Server, error) {
	server := Server{
		log:      log,
		service:  service,
		config:   config,
		listener: listener,
	}

	router := mux.NewRouter()
	router.StrictSlash(true)

	router.Handle("/", http.HandlerFunc(server.CommitTx)).Methods(http.MethodPost)

	server.server = http.Server{
		Handler: router,
	}

	return &server, nil
}

// Run starts the server that host webapp and api endpoints.
func (server *Server) Run(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	var group errgroup.Group
	group.Go(func() error {
		<-ctx.Done()
		return Error.Wrap(server.server.Shutdown(context.Background()))
	})
	group.Go(func() error {
		defer cancel()
		return Error.Wrap(server.server.Serve(server.listener))
	})

	return Error.Wrap(group.Wait())
}

// Close closes server and underlying listener.
func (server *Server) Close() error {
	return Error.Wrap(server.server.Close())
}

// CommitTx is a web api handler that is used to commit a transaction.
func (server *Server) CommitTx(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var transaction console.Transaction

	err := json.NewDecoder(r.Body).Decode(&transaction)
	if err != nil {
		server.log.Error("can not decode request body", Error.Wrap(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = server.service.CommitTx(ctx, transaction)
	if err != nil {
		server.log.Error("can not commit trasnaction", Error.Wrap(err))
		if console.ValidationError.Has(err) {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode("transaction committed successfully")
	if err != nil {
		server.log.Error("registration handler could not encode userID", Error.Wrap(err))
		return
	}
}
