package server

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/internal/handler"
)

func Start(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to bind to %s: %v", address, err)
	}
	defer listener.Close()

	fmt.Println("Server started on", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handler.HandleConnection(conn)
	}
}
