package commands

import (
	"fmt"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/models/core"
)

func HandleXread(conn net.Conn, args []string) {
	if len(args) < 4 || strings.ToLower(args[1]) != "streams" {
		conn.Write([]byte("-ERR wrong number of arguments for 'XREAD' command\r\n"))
		return
	}

	streamsAndIDs := args[2:]
	numStreams := len(streamsAndIDs) / 2
	if len(streamsAndIDs)%2 != 0 {
		conn.Write([]byte("-ERR unbalanced list of streams vs IDs\r\n"))
		return
	}

	streamKeys := streamsAndIDs[:numStreams]
	startIDs := streamsAndIDs[numStreams:]

	mu.RLock()
	defer mu.RUnlock()

	var response []interface{}
	for i, streamKey := range streamKeys {
		entry, exists := store[streamKey]
		if !exists || entry.Type != "stream" {
			continue
		}

		stream, ok := entry.Data.(core.Stream)
		if !ok {
			continue
		}

		startID := startIDs[i]
		var entries []core.StreamEntry
		for _, entry := range stream.Entries {
			compare, err := compareIDs(entry.ID, startID)
			if err != nil || compare <= 0 {
				continue
			}
			entries = append(entries, entry)
		}

		if len(entries) > 0 {
			streamResponse := []interface{}{
				streamKey,
				entries,
			}
			response = append(response, streamResponse)
		}
	}

	resp := encodeXreadResponse(response)
	conn.Write([]byte(resp))
}

func encodeXreadResponse(response []interface{}) string {
	if len(response) == 0 {
		return "*0\r\n"
	}

	var resp strings.Builder
	resp.WriteString(fmt.Sprintf("*%d\r\n", len(response)))

	for _, streamResponse := range response {
		streamData := streamResponse.([]interface{})
		streamKey := streamData[0].(string)
		entries := streamData[1].([]core.StreamEntry)

		resp.WriteString(fmt.Sprintf("*2\r\n$%d\r\n%s\r\n", len(streamKey), streamKey))
		resp.WriteString(fmt.Sprintf("*%d\r\n", len(entries)))

		for _, entry := range entries {
			resp.WriteString(fmt.Sprintf("*2\r\n$%d\r\n%s\r\n", len(entry.ID), entry.ID))

			var fields []string
			for key, value := range entry.Fields {
				fields = append(fields, key, value)
			}

			resp.WriteString(fmt.Sprintf("*%d\r\n", len(fields)))
			for _, field := range fields {
				resp.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(field), field))
			}
		}
	}

	return resp.String()
}
