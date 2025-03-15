package commands

import (
	"net"
)

func HandlePing(conn net.Conn, args []string, isReplica bool) {
	if isReplica {
		return
	}

	if conn != nil {
		conn.Write([]byte("+PONG\r\n"))
	}
}
