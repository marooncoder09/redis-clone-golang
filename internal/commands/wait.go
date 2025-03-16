package commands

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/replication"
)

func HandleWait(conn net.Conn, args []string, isReplica bool) {
	if isReplica {
		return
	}
	if len(args) < 3 {
		conn.Write([]byte("-ERR wrong number of arguments for 'WAIT' command\r\n"))
		return
	}
	numReplicas, err := strconv.Atoi(args[1])
	if err != nil {
		conn.Write([]byte("-ERR invalid numreplicas\r\n"))
		return
	}
	timeoutMs, err := strconv.Atoi(args[2])
	if err != nil {
		conn.Write([]byte("-ERR invalid timeout\r\n"))
		return
	}

	desiredOffset := replication.GetOffset()
	if replication.GetReplicaCount() == 0 {
		conn.Write([]byte(":0\r\n"))
		return
	}

	replication.RequestAckFromReplicas()

	deadline := time.Now().Add(time.Duration(timeoutMs) * time.Millisecond)
	var count int
	for {
		count = replication.CountReplicasAtOrAboveOffset(desiredOffset)
		if count >= numReplicas {
			break
		}
		if time.Now().After(deadline) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	conn.Write([]byte(fmt.Sprintf(":%d\r\n", count)))
}
