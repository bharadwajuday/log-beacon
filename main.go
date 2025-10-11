package main

import (
	"log-beacon/internal/server"
)

func main() {
	// Create a new router from our server package.
	router := server.NewRouter()

	// Start the server on port 8080.
	if err := router.Run(":8080"); err != nil {
		// Using panic is okay for now, but in a real app,
		// you'd want more graceful error handling.
		panic(err)
	}
}