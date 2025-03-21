package commands

import (
	"net"
)

func HandleExec(conn net.Conn) {
	conn.Write([]byte("-ERR EXEC without MULTI\r\n"))
}
