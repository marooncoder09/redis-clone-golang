package handler

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/commands"
	parser "github.com/codecrafters-io/redis-starter-go/internal/parser"
)

func HandleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		args, err := parser.ParseRequest(reader)
		if err != nil {
			fmt.Println("Error parsing request:", err)
			return
		}
		if len(args) == 0 {
			continue
		}
		processCommand(conn, args)
	}
}

func processCommand(conn net.Conn, args []string) {
	command := strings.ToUpper(args[0])
	fmt.Println("Processing command:", command)

	switch command {
	case "PING":
		commands.HandlePing(conn)
	case "ECHO":
		commands.HandleEcho(conn, args)
	case "SET":
		commands.HandleSet(conn, args)
	case "GET":
		commands.HandleGet(conn, args)
	case "CONFIG":
		commands.HandleConfig(conn, args)
	default:
		conn.Write([]byte(fmt.Sprintf("-ERR unknown command '%s'\r\n", command)))
	}
}
