package commands

import (
	"fmt"
	"net"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/internal/models/core"
)

func HandleIncr(conn net.Conn, args []string) {
	if len(args) < 2 {
		conn.Write([]byte("-ERR wrong number of arguments for 'INCR' command\r\n"))
		return
	}

	key := args[1]

	mu.Lock()
	defer mu.Unlock()

	entry, exists := store[key]
	if !exists {
		// Key does not exist, initialize it to 1
		store[key] = core.StoreEntry{
			Data: "1",
			Type: "string",
		}
		conn.Write([]byte(":1\r\n"))
		return
	}

	if entry.Type != "string" {
		conn.Write([]byte("-ERR value is not an integer or out of range\r\n"))
		return
	}

	strValue, ok := entry.Data.(string)
	if !ok {
		conn.Write([]byte("-ERR value is not a string\r\n"))
		return
	}

	intValue, err := strconv.Atoi(strValue)
	if err != nil {
		conn.Write([]byte("-ERR value is not an integer\r\n"))
		return
	}

	intValue++
	newStrValue := strconv.Itoa(intValue)

	store[key] = core.StoreEntry{
		Data:      newStrValue,
		Type:      "string",
		ExpiresAt: entry.ExpiresAt,
	}

	conn.Write([]byte(fmt.Sprintf(":%d\r\n", intValue)))
}
