package commands

import (
	"encoding/hex"
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/internal/replication"
	"github.com/codecrafters-io/redis-starter-go/internal/utils"
)

const emptyRDBHex = "524544495330303036ff00006b2471d24e010000" // hardcoded rn, but pick this from the .env file instead

func HandlePsync(conn net.Conn, args []string) {
	if len(args) != 3 {
		conn.Write([]byte("-ERR invalid PSYNC command\r\n"))
		return
	}

	masterReplID, exists := GetConfig("master_replid")
	if !exists {
		masterReplID = utils.GetMasterReplID()
	}

	fullResyncResponse := fmt.Sprintf("+FULLRESYNC %s 0\r\n", masterReplID)
	conn.Write([]byte(fullResyncResponse))

	rdbBytes, err := hex.DecodeString(emptyRDBHex)
	if err != nil {
		conn.Write([]byte("-ERR failed to load empty RDB file\r\n"))
		return
	}

	rdbHeader := fmt.Sprintf("$%d\r\n", len(rdbBytes))
	conn.Write([]byte(rdbHeader))
	conn.Write(rdbBytes)

	replication.AddReplica(conn)
}
