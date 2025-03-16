package handler

import (
	"bufio"
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/internal/commands"
	parser "github.com/codecrafters-io/redis-starter-go/internal/parser"
)

func HandleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		args, _, err := parser.ParseRequestWithByteCount(reader)
		if err != nil {
			fmt.Println("Error parsing request:", err)
			return
		}
		if len(args) == 0 {
			continue
		}
		commands.ProcessCommand(conn, args, false)

	}
}
