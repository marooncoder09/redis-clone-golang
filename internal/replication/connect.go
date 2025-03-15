package replication

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/parser"
)

type CommandHandlerFunc func(conn net.Conn, args []string, isReplica bool)

func StartReplicaProcess(masterAddress, replicaPort string, handleCommand CommandHandlerFunc) {
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

	reader := bufio.NewReader(conn)

	fmt.Println("Connected to master. Starting handshake...")
	performHandshake(conn, reader, replicaPort)

	go HandleReplicatedCommands(reader, handleCommand)
}

func HandleReplicatedCommands(reader *bufio.Reader, handleCommand CommandHandlerFunc) {
	for {
		args, err := parser.ParseRequest(reader)
		if err != nil {
			if err == io.EOF || strings.Contains(err.Error(), "closed network connection") {
				fmt.Println("[Replica] Master disconnected")
				return
			}
			fmt.Println("[Replica] Error parsing replicated command:", err)
			return
		}
		if len(args) == 0 {
			continue
		}

		fmt.Println("[Replica] Received command from master:", args)
		handleCommand(nil, args, true)
	}
}
