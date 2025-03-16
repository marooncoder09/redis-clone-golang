package commands

import (
	"fmt"
	"net"
	"strings"
)

func ProcessCommand(conn net.Conn, args []string, isReplica bool) {
	command := strings.ToUpper(args[0])
	fmt.Println("Processing command:", command)

	switch command {
	case "PING":
		HandlePing(conn, args, isReplica)
	case "ECHO":
		HandleEcho(conn, args)
	case "SET":
		HandleSet(conn, args, isReplica)
	case "GET":
		HandleGet(conn, args)
	case "CONFIG":
		HandleConfig(conn, args)
	case "KEYS":
		HandleKeys(conn, args)
	case "INFO":
		HandleInfo(conn, args)
	case "REPLCONF":
		HandleReplConf(conn, args)
	case "PSYNC":
		HandlePsync(conn, args)
	case "WAIT":
		HandleWait(conn, args, isReplica)
	case "TYPE":
		HandleType(conn, args)
	default:
		conn.Write([]byte(fmt.Sprintf("-ERR unknown command '%s'\r\n", command)))
	}
}
