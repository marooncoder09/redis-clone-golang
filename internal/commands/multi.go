package commands

import (
	"net"
)

func HandleMulti(conn net.Conn) {
	conn.Write([]byte("+OK\r\n"))
}
