package commands

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/internal/utils"
)

func HandlePsync(conn net.Conn, args []string) {
	if len(args) != 3 {
		conn.Write([]byte("-ERR invalid PSYNC command\r\n"))
		return
	}

	masterReplID, exists := GetConfig("master_replid")
	if !exists {
		masterReplID = utils.GetMasterReplID()
	}

	response := fmt.Sprintf("+FULLRESYNC %s 0\r\n", masterReplID)
	conn.Write([]byte(response))
}
