package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/asherzog/thisor/internal/router"
)

type options struct {
	lg         *slog.Logger
	listenAddr string
}

func runHttp(opts options) error {
	router, err := router.NewRouter(opts.lg)
	if err != nil {
		opts.lg.Error("Failed to initialize router")
		return err
	}

	s := http.Server{
		Addr:         opts.listenAddr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	opts.lg.Info("Starting Server", "addr:", opts.listenAddr)
	return s.ListenAndServe()
}
