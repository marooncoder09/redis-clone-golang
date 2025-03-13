package replication

import (
	"fmt"
	"log"
	"net"
	"strings"
)

func StartReplicaProcess(masterAddress, replicaPort string) {
	parts := strings.Split(masterAddress, " ")
	if len(parts) != 2 {
		log.Println("Invalid --replicaof format. Expected: '<MASTER_HOST> <MASTER_PORT>'")
		return
	}

	host, port := parts[0], parts[1]
	address := fmt.Sprintf("%s:%s", host, port)

	fmt.Println("Connecting to master at", address)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Println("Failed to connect to master:", err)
		return
	}

	fmt.Println("Connected to master. Starting handshake...")

	performHandshake(conn, replicaPort)
}
