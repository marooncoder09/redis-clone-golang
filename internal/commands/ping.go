package commands

import (
	"fmt"
	"net"
)

func HandlePing(conn net.Conn) {
	conn.Write([]byte("+PONG\r\n"))
	fmt.Println("Processed PING command")
}
