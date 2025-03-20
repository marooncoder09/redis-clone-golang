package commands

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/models/core"
)

func HandleXrange(conn net.Conn, args []string) {
	if len(args) < 4 {
		conn.Write([]byte("-ERR wrong number of arguments for 'XRANGE' command\r\n"))
		return
	}

	streamKey := args[1]
	startID := args[2]
	endID := args[3]

	mu.RLock()
	entry, exists := store[streamKey]
	mu.RUnlock()

	if !exists || entry.Type != "stream" {
		conn.Write([]byte("*0\r\n")) // returns an empty array if the stream doesn't exist
		return
	}

	stream, ok := entry.Data.(core.Stream)
	if !ok {
		conn.Write([]byte("-ERR invalid stream data\r\n"))
		return
	}

	var result []core.StreamEntry
	for _, entry := range stream.Entries {
		if startID != "-" {
			compareStart, err := compareIDs(entry.ID, startID)
			if err != nil {
				continue
			}
			if compareStart < 0 {
				continue
			}
		}

		compareEnd, err := compareIDs(entry.ID, endID)
		if err != nil {
			continue
		}
		if compareEnd > 0 {
			continue
		}

		result = append(result, entry)
	}

	resp := encodeXrangeResponse(result)
	conn.Write([]byte(resp))
}

func encodeXrangeResponse(entries []core.StreamEntry) string {
	var resp strings.Builder
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

	return resp.String()
}

func parseIDForXrange(id string) (int64, int64, error) {
	parts := strings.Split(id, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid ID format")
	}

	millisecondsTime, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid millisecondsTime")
	}

	sequenceNumber, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid sequenceNumber")
	}

	return millisecondsTime, sequenceNumber, nil
}

// compareIDs compares two IDs and returns:
// -1 if id1 < id2
// 0 if id1 == id2
// 1 if id1 > id2
func compareIDs(id1, id2 string) (int, error) {
	if id1 == "-" {
		return -1, nil
	}
	if id2 == "-" {
		return 1, nil
	}

	if !strings.Contains(id1, "-") {
		id1 += "-0"
	}
	if !strings.Contains(id2, "-") {
		id2 += "-0"
	}

	millis1, seq1, err := parseIDForXrange(id1)
	if err != nil {
		return 0, err
	}

	millis2, seq2, err := parseIDForXrange(id2)
	if err != nil {
		return 0, err
	}

	if millis1 < millis2 {
		return -1, nil
	} else if millis1 > millis2 {
		return 1, nil
	} else {
		if seq1 < seq2 {
			return -1, nil
		} else if seq1 > seq2 {
			return 1, nil
		} else {
			return 0, nil
		}
	}
}
