package commands

import (
	"fmt"
	"net"
	"strconv"
)

func HandleWait(conn net.Conn, args []string, isReplica bool) {
	if len(args) < 3 {
		if conn != nil {
			conn.Write([]byte("-ERR wrong number of arguments for 'WAIT' command\r\n"))
		}
		return
	}

	numReplicasStr := args[1]
	_, err := strconv.Atoi(numReplicasStr)
	if err != nil {
		if conn != nil {
			conn.Write([]byte("-ERR invalid numreplicas\r\n"))
		}
		return
	}
	timeoutStr := args[2]
	_, err = strconv.Atoi(timeoutStr)
	if err != nil {
		if conn != nil {
			conn.Write([]byte("-ERR invalid timeout\r\n"))
		}
		return
	}

	if !isReplica {
		reply := fmt.Sprintf(":%d\r\n", 0)
		conn.Write([]byte(reply))
	}
}
