package commands

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/internal/models/core"
)

func HandleXadd(conn net.Conn, args []string) {
	if len(args) < 4 {
		conn.Write([]byte("-ERR wrong number of arguments for 'XADD' command\r\n"))
		return
	}

	streamKey := args[1]
	entryID := args[2]
	fields := args[3:]

	if len(fields)%2 != 0 {
		conn.Write([]byte("-ERR wrong number of arguments for fields\r\n"))
		return
	}

	fieldMap := make(map[string]string)
	for i := 0; i < len(fields); i += 2 {
		field := fields[i]
		value := fields[i+1]
		fieldMap[field] = value
	}

	mu.Lock()
	defer mu.Unlock()

	entry, exists := store[streamKey]
	if !exists {
		newStream := core.Stream{
			Entries: []core.StreamEntry{
				{
					ID:     entryID,
					Fields: fieldMap,
				},
			},
		}
		store[streamKey] = core.StoreEntry{
			Type: "stream",
			Data: newStream,
		}
	} else {
		if entry.Type != "stream" {
			conn.Write([]byte("-ERR key exists and is not a stream\r\n"))
			return
		}

		stream, ok := entry.Data.(core.Stream)
		if !ok {
			conn.Write([]byte("-ERR invalid stream data\r\n"))
			return
		}

		stream.Entries = append(stream.Entries, core.StreamEntry{
			ID:     entryID,
			Fields: fieldMap,
		})
		entry.Data = stream
		store[streamKey] = entry
	}

	resp := fmt.Sprintf("$%d\r\n%s\r\n", len(entryID), entryID)
	conn.Write([]byte(resp))
}
