package main

import (
	"flag"
	"log"
	"log/slog"
	"net"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if os.Getenv("IS_LOCAL") == "true" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}
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
