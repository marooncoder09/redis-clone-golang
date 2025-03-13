package commands

import (
	"fmt"
	"net"
	"strings"
)

func HandleInfo(conn net.Conn, args []string) {
	if len(args) < 2 || strings.ToLower(args[1]) != "replication" {
		conn.Write([]byte("-ERR unsupported INFO section\r\n"))
		return
	}

	role, exists := GetConfig("role")
	if !exists {
		role = "master"
	}

	infoResponse := fmt.Sprintf("role:%s", role)
	response := fmt.Sprintf("$%d\r\n%s\r\n", len(infoResponse), infoResponse)

	conn.Write([]byte(response))
}
