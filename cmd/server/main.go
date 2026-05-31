package main

import (
	"log"

	"mini-card-game/internal/server"
)

func main() {
	r, addr, cleanup, err := server.New(server.Options{})
	if err != nil {
		log.Fatalf("init server failed: %v", err)
	}
	defer cleanup()

	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
