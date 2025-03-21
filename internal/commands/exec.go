package commands

import (
	"net"

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
	models.ClientMu.Unlock()

	conn.Write([]byte("*0\r\n"))
}
