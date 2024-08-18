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
	s := http.Server{
		Addr:         opts.listenAddr,
		Handler:      router.NewRouter(opts.lg),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	opts.lg.Info("Starting Server", "addr:", opts.listenAddr)
	return s.ListenAndServe()
}
