package commands

import (
	"fmt"
	"net"
	"strings"

	models "github.com/codecrafters-io/redis-starter-go/internal/models/core"
)

func ProcessCommand(conn net.Conn, args []string, isReplica bool) {
	command := strings.ToUpper(args[0])
	fmt.Println("Processing command:", command)

	models.ClientMu.Lock()
	state, exists := models.ClientStates[conn]
	if !exists {
		state = &models.ClientState{
			InTransaction: false,
			CommandQueue:  make([][]string, 0),
		}
		models.ClientStates[conn] = state
	}
	inTransaction := state.InTransaction
	models.ClientMu.Unlock()

	// If we're inside MULTI (transaction), we will queue all commands
	// except MULTI and EXEC themselves
	if inTransaction && command != "MULTI" && command != "EXEC" {
		models.ClientMu.Lock()
		state.CommandQueue = append(state.CommandQueue, args)
		models.ClientMu.Unlock()

		conn.Write([]byte("+QUEUED\r\n"))
		return
	}

	// otherwise we will execute the commands normally
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
	case "XADD":
		HandleXadd(conn, args)
	case "XRANGE":
		HandleXrange(conn, args)
	case "XREAD":
		HandleXread(conn, args)
	case "INCR":
		HandleIncr(conn, args)
	case "MULTI":
		HandleMulti(conn)
	case "EXEC":
		HandleExec(conn)
	default:
		conn.Write([]byte(fmt.Sprintf("-ERR unknown command '%s'\r\n", command)))
	}
}
