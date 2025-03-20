package commands

import (
	"fmt"
	"net"
	"strconv"
	"strings"

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

	// Check if the entry ID is in the format "millis-*"
	parts := strings.Split(entryID, "-")
	if len(parts) == 2 && parts[1] == "*" {
		millisPart := parts[0]
		newMillis, err := strconv.ParseInt(millisPart, 10, 64)
		if err != nil {
			conn.Write([]byte("-ERR invalid milliseconds part\r\n"))
			return
		}

		var seq int64
		entry, exists := store[streamKey]
		var stream core.Stream

		if exists {
			if entry.Type != "stream" {
				conn.Write([]byte("-ERR key exists and is not a stream\r\n"))
				return
			}
			stream = entry.Data.(core.Stream)
		}

		if exists && len(stream.Entries) > 0 {
			lastEntryID := stream.Entries[len(stream.Entries)-1].ID
			lastMillis, lastSeq, err := parseID(lastEntryID)
			if err != nil {
				conn.Write([]byte("-ERR invalid last entry ID\r\n"))
				return
			}

			if newMillis < lastMillis {
				conn.Write([]byte("-ERR The ID specified in XADD is equal or smaller than the target stream top item\r\n"))
				return
			} else if newMillis == lastMillis {
				seq = lastSeq + 1
			} else {
				seq = 0
			}
		} else {
			// Stream is empty
			if newMillis == 0 {
				seq = 1
			} else {
				seq = 0
			}
		}

		newID := fmt.Sprintf("%d-%d", newMillis, seq)
		// Validate the newID
		if exists && len(stream.Entries) > 0 {
			lastEntryID := stream.Entries[len(stream.Entries)-1].ID
			if err := validateID(newID, lastEntryID); err != nil {
				conn.Write([]byte(fmt.Sprintf("-%s\r\n", err.Error())))
				return
			}
		} else {
			if err := validateID(newID, ""); err != nil {
				conn.Write([]byte(fmt.Sprintf("-%s\r\n", err.Error())))
				return
			}
		}

		entryID = newID
		args[2] = newID // Update args for potential propagation
	}

	entry, exists := store[streamKey]
	if !exists {
		// stream is empty, time to validate the new id
		if err := validateID(entryID, ""); err != nil {
			conn.Write([]byte(fmt.Sprintf("-%s\r\n", err.Error())))
			return
		}

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

		lastEntryID := ""
		if len(stream.Entries) > 0 {
			lastEntryID = stream.Entries[len(stream.Entries)-1].ID
		}

		if err := validateID(entryID, lastEntryID); err != nil {
			conn.Write([]byte(fmt.Sprintf("-%s\r\n", err.Error())))
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

func parseID(id string) (int64, int64, error) {
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

func validateID(newID string, lastEntryID string) error {
	if newID == "0-0" {
		return fmt.Errorf("ERR The ID specified in XADD must be greater than 0-0")
	}

	newMillis, newSeq, err := parseID(newID)
	if err != nil {
		return err
	}

	if lastEntryID == "" {
		// the stream is empty, new ids must be greater than 0-0, and if the stream is not empty then the stream id myst be greater then last entry's id
		if newMillis < 0 || (newMillis == 0 && newSeq < 1) || (newMillis > 0 && newSeq < 0) {
			return fmt.Errorf("ERR The ID specified in XADD must be greater than 0-0")
		}
		return nil
	}

	lastMillis, lastSeq, err := parseID(lastEntryID)
	if err != nil {
		return err
	}

	if newMillis < lastMillis || (newMillis == lastMillis && newSeq <= lastSeq) {
		return fmt.Errorf("ERR The ID specified in XADD is equal or smaller than the target stream top item")
	}

	return nil
}
