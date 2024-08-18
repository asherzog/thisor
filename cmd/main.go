package main

import (
	"flag"
	"log/slog"
	"net"
	"os"
)

func main() {
	var (
		host = flag.String("host", "", "host http address to listen on")
		port = flag.String("port", "8080", "port number for http listener")
	)
	flag.Parse()

	addr := net.JoinHostPort(*host, *port)
	lg := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := runHttp(options{
		lg:         lg,
		listenAddr: addr,
	}); err != nil {
		lg.Error(err.Error())
		return
	}
}
