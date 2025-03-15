package commands

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/replication"
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

	case "getack":
		offsetValue := replication.GetOffset()

		offsetStr := strconv.FormatInt(offsetValue, 10)
		ackResponse := fmt.Sprintf(
			"*3\r\n$8\r\nREPLCONF\r\n$3\r\nACK\r\n$%d\r\n%s\r\n",
			len(offsetStr),
			offsetStr,
		)
		conn.Write([]byte(ackResponse))
		return

	default:
		conn.Write([]byte("-ERR unknown REPLCONF parameter\r\n"))
		return
	}

	conn.Write([]byte("+OK\r\n"))
}
