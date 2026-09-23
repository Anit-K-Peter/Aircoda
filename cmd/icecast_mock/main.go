package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"aircoda/internal/audio"
)

func main() {
	configPath := flag.String("c", "", "Path to Icecast XML config")
	flag.Parse()

	if *configPath == "" {
		fmt.Println("Usage: icecast -c <config.xml>")
		os.Exit(1)
	}

	server := audio.NewStandaloneIcecastServer("/986fm")
	if err := server.ListenAndServe("127.0.0.1", 8000); err != nil {
		fmt.Fprintf(os.Stderr, "Icecast server error: %v\n", err)
		os.Exit(1)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	server.Stop()
}
