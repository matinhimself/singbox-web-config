package main

import (
	"flag"
	"log"

	"github.com/matinhimself/singbox-web-config/internal/handlers"
)

func main() {
	addr := flag.String("addr", "localhost:8080", "HTTP server address")
	flag.Parse()

	server, err := handlers.NewServer(*addr)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	log.Printf("Sing-Box Config Manager")
	log.Printf("=======================")

	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
