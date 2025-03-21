package commands

import (
	"net"

	models "github.com/codecrafters-io/redis-starter-go/internal/models/core"
)

func HandleDiscard(conn net.Conn) {
	models.ClientMu.Lock()
	defer models.ClientMu.Unlock()

	state, exists := models.ClientStates[conn]
	if !exists || !state.InTransaction {
		conn.Write([]byte("-ERR DISCARD without MULTI\r\n"))
		return
	}

	// reset transaction state and clear the command queue
	state.InTransaction = false
	state.CommandQueue = make([][]string, 0)

	conn.Write([]byte("+OK\r\n"))
}
