package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.Println("Hot-storage consumer started. (Placeholder - No-op)")

	// Wait for a termination signal to gracefully shut down.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	log.Println("Shutting down hot-storage consumer...")
}
