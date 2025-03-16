package utils

import (
	"fmt"
	"net"
)

func SendError(conn net.Conn, message string) {
	conn.Write([]byte(fmt.Sprintf("-ERR %s\r\n", message)))
}
