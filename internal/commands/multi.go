package commands

import (
	"net"

	models "github.com/codecrafters-io/redis-starter-go/internal/models/core"
)

func HandleMulti(conn net.Conn) {
	models.ClientMu.Lock()
	defer models.ClientMu.Unlock()

	state, exists := models.ClientStates[conn]
	if !exists {
		state = &models.ClientState{InTransaction: true}
		models.ClientStates[conn] = state
	} else {
		state.InTransaction = true
	}

	conn.Write([]byte("+OK\r\n"))
}
