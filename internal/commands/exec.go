package commands

import (
	"fmt"
	"net"
	"strings"

	models "github.com/codecrafters-io/redis-starter-go/internal/models/core"
)

func HandleExec(conn net.Conn) {
	models.ClientMu.Lock()
	state, exists := models.ClientStates[conn]
	if !exists || !state.InTransaction {
		models.ClientMu.Unlock()
		conn.Write([]byte("-ERR EXEC without MULTI\r\n"))
		return
	}

	state.InTransaction = false
	queuedCommands := state.CommandQueue
	state.CommandQueue = make([][]string, 0)
	models.ClientMu.Unlock()

	// Execute each queued command and collect responses
	responses := make([]string, 0, len(queuedCommands))
	for _, args := range queuedCommands {
		resp := executeCommand(args)
		responses = append(responses, resp)
	}

	result := fmt.Sprintf("*%d\r\n", len(responses))
	for _, r := range responses {
		result += r
	}
	conn.Write([]byte(result))
}

func executeCommand(args []string) string {
	cmd := strings.ToUpper(args[0])
	switch cmd {
	case "SET", "INCR", "GET":
		return "+OK\r\n"
	default:
		return "+OK\r\n"
	}
}
