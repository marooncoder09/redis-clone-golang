package commands

import (
	"fmt"
	"net"
)

func HandleEcho(conn net.Conn, args []string) {
	if len(args) > 1 {
		response := args[1]
		conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(response), response)))
	} else {
		conn.Write([]byte("$0\r\n\r\n"))
	}
	fmt.Println("Processed ECHO command:", args)
}
