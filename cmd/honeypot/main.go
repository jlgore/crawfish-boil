package main

import (
	"flag"
	"log"

	"openclaw-honeypot/internal/gateway"
	"openclaw-honeypot/internal/logging"
)

var version = "dev"

func main() {
	addr := flag.String("addr", ":18789", "listen address")
	flag.Parse()

	logging.Init()

	srv := gateway.NewServer(*addr)
	log.Fatal(srv.ListenAndServe())
}
