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
	entry, exists := GetEntry(key)

	response := "+none\r\n"
	if exists {
		switch entry.Type {
		case "stream":
			response = "+stream\r\n"
		case "string":
			response = "+string\r\n"
		default:
			response = "+none\r\n"
		}
	}

	conn.Write([]byte(response))
}
