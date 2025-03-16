package commands

import (
	"net"

	"github.com/codecrafters-io/redis-starter-go/internal/utils"
)

func HandleType(conn net.Conn, args []string) {
	if len(args) < 2 {
		utils.SendError(conn, "wrong number of arguments for 'TYPE' command")
		return
	}

	key := args[1]
	_, exists := GetKey(key)

	response := "+none\r\n"
	if exists {
		response = "+string\r\n"
	}

	conn.Write([]byte(response))
}
