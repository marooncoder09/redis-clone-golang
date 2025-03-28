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

	go HandleReplicatedCommands(conn, reader, handleCommand)
}

func HandleReplicatedCommands(
	conn net.Conn,
	reader *bufio.Reader,
	handleCommand CommandHandlerFunc,
) {

	SetOffset(0)

	for {

		args, nBytes, err := parser.ParseRequestWithByteCount(reader)
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

		if len(args) >= 2 && strings.ToUpper(args[0]) == "REPLCONF" && strings.ToUpper(args[1]) == "GETACK" {
			handleCommand(conn, args, true)
			AddToOffset(nBytes)
		} else {
			AddToOffset(nBytes)
			handleCommand(conn, args, true)
		}
	}
}
