package commands

import (
	"fmt"
	"net"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/internal/replication"
)

func HandleWait(conn net.Conn, args []string, isReplica bool) {
	if len(args) < 3 {
		if conn != nil && !isReplica {
			conn.Write([]byte("-ERR wrong number of arguments for 'WAIT' command\r\n"))
		}
		return
	}

	if isReplica {
		return
	}

	_, err := strconv.Atoi(args[1])
	if err != nil {
		conn.Write([]byte("-ERR invalid numreplicas\r\n"))
		return
	}
	_, err = strconv.Atoi(args[2])
	if err != nil {
		conn.Write([]byte("-ERR invalid timeout\r\n"))
		return
	}

	count := replication.GetReplicaCount()

	response := fmt.Sprintf(":%d\r\n", count)
	conn.Write([]byte(response))
}
