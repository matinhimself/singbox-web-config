package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/matinhimself/singbox-web-config/internal/handlers"
)

func main() {
	addr := flag.String("addr", "localhost:8080", "HTTP server address")
	configPath := flag.String("config", "/etc/sing-box/config.json", "Path to sing-box config file")
	serviceName := flag.String("service", "sing-box", "Name of sing-box systemd service")
	flag.Parse()

	log.Printf("Sing-Box Config Manager")
	log.Printf("=======================")
	log.Printf("Config path: %s", *configPath)
	log.Printf("Service name: %s", *serviceName)
	log.Printf("")

	server, err := handlers.NewServer(*addr, *configPath, *serviceName)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\nShutting down gracefully...")
		server.Stop()
		os.Exit(0)
	}()

	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
