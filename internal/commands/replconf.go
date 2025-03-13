package commands

import (
	"fmt"
	"net"
	"strings"
)

func HandleReplConf(conn net.Conn, args []string) {
	if len(args) < 3 {
		conn.Write([]byte("-ERR invalid REPLCONF command\r\n"))
		return
	}

	param := strings.ToLower(args[1])

	switch param {
	case "listening-port":

		if len(args) != 3 {
			conn.Write([]byte("-ERR missing port argument\r\n"))
			return
		}
		fmt.Println("Replica reported listening port:", args[2])

	case "capa":

		if len(args) != 3 {
			conn.Write([]byte("-ERR missing capa argument\r\n"))
			return
		}
		fmt.Println("Replica supports capability:", args[2])

	default:
		conn.Write([]byte("-ERR unknown REPLCONF parameter\r\n"))
		return
	}

	conn.Write([]byte("+OK\r\n"))
}
