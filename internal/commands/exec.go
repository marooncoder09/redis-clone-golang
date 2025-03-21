package commands

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/models/core"
	models "github.com/codecrafters-io/redis-starter-go/internal/models/core"
	"github.com/codecrafters-io/redis-starter-go/internal/replication"
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

	responses := make([]string, 0, len(queuedCommands))
	for _, args := range queuedCommands {
		response := executeCommand(conn, args)
		responses = append(responses, response)
	}

	result := fmt.Sprintf("*%d\r\n", len(responses))
	for _, r := range responses {
		result += r
	}
	conn.Write([]byte(result))
}

func executeCommand(conn net.Conn, args []string) string {
	cmd := strings.ToUpper(args[0])
	switch cmd {
	case "SET":
		return handleSetInTransaction(conn, args)
	case "INCR":
		return handleIncrInTransaction(conn, args)
	case "GET":
		return handleGetInTransaction(conn, args)
	default:
		return "+OK\r\n"
	}
}

func handleSetInTransaction(conn net.Conn, args []string) string {
	if len(args) < 3 {
		return "-ERR wrong number of arguments for 'SET' command\r\n"
	}

	key := args[1]
	value := args[2]
	var ttl int64 = 0

	if len(args) >= 5 && strings.ToUpper(args[3]) == "PX" {
		parsedTTL, err := strconv.ParseInt(args[4], 10, 64)
		if err != nil || parsedTTL <= 0 {
			return "-ERR PX must be a positive integer\r\n"
		}
		ttl = parsedTTL
	}

	SetKey(key, value, ttl)
	replication.PropagateCommand("SET", args[1:])
	return "+OK\r\n"
}

func handleIncrInTransaction(conn net.Conn, args []string) string {
	if len(args) < 2 {
		return "-ERR wrong number of arguments for 'INCR' command\r\n"
	}

	key := args[1]
	entry, exists := GetEntry(key)
	if !exists {
		entry = core.StoreEntry{Data: "0", Type: "string"}
	}

	val, err := strconv.Atoi(entry.Data.(string))
	if err != nil {
		return "-ERR value is not an integer\r\n"
	}

	val++
	SetKey(key, strconv.Itoa(val), 0)
	replication.PropagateCommand("INCR", args[1:])
	return fmt.Sprintf(":%d\r\n", val)
}

func handleGetInTransaction(conn net.Conn, args []string) string {
	if len(args) < 2 {
		return "-ERR wrong number of arguments for 'GET' command\r\n"
	}

	key := args[1]
	entry, exists := GetEntry(key)
	if !exists {
		return "$-1\r\n"
	}

	value, ok := entry.Data.(string)
	if !ok {
		return "$-1\r\n"
	}

	return fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
}
