package commands

import (
	"fmt"
	"log"
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

		replication.AddReplica(conn)
		conn.Write([]byte("+OK\r\n"))
		return

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

		if _, err := conn.Write([]byte(ackResponse)); err != nil {
			log.Println("Failed to send ACK:", err)
		}
		return

	case "ack":
		if len(args) < 3 {
			conn.Write([]byte("-ERR missing offset\r\n"))
			return
		}
		ackOffset, _ := strconv.ParseInt(args[2], 10, 64)
		replication.SetReplicaOffset(conn, ackOffset)
		return

	default:
		conn.Write([]byte("-ERR unknown REPLCONF parameter\r\n"))
		return
	}

	conn.Write([]byte("+OK\r\n"))
}
