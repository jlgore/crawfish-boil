package main

import (
	"flag"
	"log"
	"os"

	"openclaw-honeypot/internal/gateway"
	"openclaw-honeypot/internal/geoip"
	"openclaw-honeypot/internal/logging"
)

var version = "dev"

func main() {
	addr := flag.String("addr", ":18789", "listen address")
	flag.Parse()

	logging.Init()

	geo := geoip.NewClient(os.Getenv("IPINFO_TOKEN"))

	srv := gateway.NewServer(*addr, geo)
	log.Fatal(srv.ListenAndServe())
}
