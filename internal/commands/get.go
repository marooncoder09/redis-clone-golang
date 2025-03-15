package commands

import (
	"fmt"
	"net"
)

func HandleGet(conn net.Conn, args []string) {
	if len(args) < 2 {
		conn.Write([]byte("-ERR wrong number of arguments for 'GET' command\r\n"))
		return
	}

	key := args[1]
	value, exists := GetKey(key)

	if !exists {
		conn.Write([]byte("$-1\r\n")) // Null bulk string if key expired
		fmt.Println("Processed GET:", key, "-> Expired or not found")
		return
	}

	resp := fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
	conn.Write([]byte(resp))
}
