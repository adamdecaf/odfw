// Copyright 2020 Adam Shannon
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adamdecaf/odfw/internal/exec"

	"github.com/gorilla/mux"
)

var (
	flagBasePath = flag.String("http.base-path", "", "Optional path to serve all requests from")
	flagHttpAddr = flag.String("http.addr", ":8888", "Bind address for stargazers HTTP server")

	flagDebug = flag.Bool("debug", false, "Enable debug logging")
)

func main() {
	flag.Parse()

	log.Printf("starting odfw %s", Version)

	shutdownCtx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	errs := make(chan error)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-shutdownCtx.Done():
				log.Print("INFO: shutting down refresh from context close")
				errs <- errors.New("shutdown via context")
				return

			case <-sig:
				log.Print("INFO: shutting down refresh from signal")
				errs <- errors.New("shutdown via signal")
				return
			}
		}
	}()

	cfg := &exec.Config{}
	exec.Basic(shutdownCtx, cfg)

	handler := mux.NewRouter()
	if *flagBasePath != "" {
		handler = handler.PathPrefix(*flagBasePath).Subrouter()
	}
	addPingRoute(handler)

	serve := &http.Server{
		Addr:    *flagHttpAddr,
		Handler: handler,
		TLSConfig: &tls.Config{
			InsecureSkipVerify:       false,
			PreferServerCipherSuites: true,
			MinVersion:               tls.VersionTLS12,
		},
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	shutdownServer := func() {
		if err := serve.Shutdown(context.TODO()); err != nil {
			log.Printf("INFO: http server shutdown: %v", err)
			errs <- err
		}
	}
	defer shutdownServer()

	go func() {
		log.Printf("HTTP server bind to %s", *flagHttpAddr)
		if err := serve.ListenAndServe(); err != nil {
			log.Printf("INFO: http server bind: %v", err)
			errs <- err
		}
	}()

	if err := <-errs; err != nil {
		log.Printf("INFO: shutdown error: %v", err)
		os.Exit(1)
	}
}

func addPingRoute(r *mux.Router) {
	r.Methods("GET").Path("/ping").HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("PONG"))
	})
}
