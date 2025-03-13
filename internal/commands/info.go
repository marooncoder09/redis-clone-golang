package commands

import (
	"fmt"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/utils"
)

func HandleInfo(conn net.Conn, args []string) {
	if len(args) < 2 || strings.ToLower(args[1]) != "replication" {
		conn.Write([]byte("-ERR unsupported INFO section\r\n"))
		return
	}

	role, _ := GetConfig("role")

	masterReplID, exists := GetConfig("master_replid")
	if !exists {
		masterReplID = utils.GetMasterReplID()
	}

	masterReplOffset, exists := GetConfig("master_repl_offset")
	if !exists {
		masterReplOffset = "0"
	}

	infoResponse := fmt.Sprintf("role:%s\r\nmaster_replid:%s\r\nmaster_repl_offset:%s", role, masterReplID, masterReplOffset)
	response := fmt.Sprintf("$%d\r\n%s\r\n", len(infoResponse), infoResponse)

	conn.Write([]byte(response))
}
