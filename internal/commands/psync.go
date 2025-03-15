package commands

import (
	"bufio"
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

	bw := bufio.NewWriter(conn)

	masterReplID := utils.GetMasterReplID()
	fullResyncResponse := fmt.Sprintf("+FULLRESYNC %s 0\r\n", masterReplID)
	if _, err := bw.WriteString(fullResyncResponse); err != nil {
		conn.Close()
		return
	}
	if err := bw.Flush(); err != nil {
		conn.Close()
		return
	}

	rdbBytes, err := hex.DecodeString(emptyRDBHex)
	if err != nil {
		conn.Write([]byte("-ERR failed to load empty RDB file\r\n"))
		conn.Close()
		return
	}

	rdbHeader := fmt.Sprintf("$%d\r\n", len(rdbBytes))
	if _, err := bw.WriteString(rdbHeader); err != nil {
		conn.Close()
		return
	}

	if _, err := bw.Write(rdbBytes); err != nil {
		conn.Close()
		return
	}

	if err := bw.Flush(); err != nil {
		conn.Close()
		return
	}

	replication.AddReplica(conn)
	fmt.Println("[Master] Registered new replica:", conn.RemoteAddr())

	go replication.HandleReplicatedCommands(bufio.NewReader(conn), ProcessCommand)
}
