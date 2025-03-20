package commands

import (
	"fmt"
	"net"
	"time"
)

func HandleGet(conn net.Conn, args []string) {
	if len(args) < 2 {
		conn.Write([]byte("-ERR wrong number of arguments for 'GET' command\r\n"))
		return
	}

	key := args[1]
	entry, exists := GetEntry(key)

	if !exists {
		conn.Write([]byte("$-1\r\n")) // key DNE
		fmt.Println("Processed GET:", key, "-> Expired or not found")
		return
	}

	if entry.ExpiresAt > 0 && time.Now().UnixMilli() > entry.ExpiresAt {
		conn.Write([]byte("$-1\r\n")) // Key is expired
		fmt.Println("Processed GET:", key, "-> Expired")
		return
	}

	if entry.Type != "string" {
		conn.Write([]byte("$-1\r\n")) // Key is not of type string
		fmt.Println("Processed GET:", key, "-> Not a string")
		return
	}

	value, ok := entry.Data.(string)
	if !ok {
		conn.Write([]byte("$-1\r\n")) // invalid data type
		fmt.Println("Processed GET:", key, "-> Invalid data type")
		return
	}

	resp := fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
	conn.Write([]byte(resp))
}
