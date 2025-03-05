package commands

import (
	"fmt"
	"net"
)

func HandleConfig(conn net.Conn, args []string) {
	if len(args) != 3 || args[1] != "GET" {
		conn.Write([]byte("-ERR syntax error\r\n"))
		return
	}

	key := args[2]
	value, exists := GetConfig(key)
	if !exists {
		conn.Write([]byte("$-1\r\n"))
		return
	}

	// RESP Array: *2\r\n$len(key)\r\nkey\r\n$len(value)\r\nvalue\r\n
	resp := fmt.Sprintf("*2\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(key), key, len(value), value)
	conn.Write([]byte(resp))
}
