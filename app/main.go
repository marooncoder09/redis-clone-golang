package main

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/internal/server"
	"log"
)

func main() {
	fmt.Println("Starting server on port 6379...")
	err := server.Start("0.0.0.0:6379")
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
