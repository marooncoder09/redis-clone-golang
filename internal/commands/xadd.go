package commands

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

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
	var stream core.Stream

	if exists {
		if entry.Type != "stream" {
			conn.Write([]byte("-ERR key exists and is not a stream\r\n"))
			return
		}
		stream = entry.Data.(core.Stream)
	} else {
		stream = core.Stream{
			Entries: []core.StreamEntry{},
		}
	}

	if strings.HasSuffix(entryID, "-*") {
		millisPart := strings.TrimSuffix(entryID, "-*")
		millis, err := strconv.ParseInt(millisPart, 10, 64)
		if err != nil {
			conn.Write([]byte("-ERR invalid milliseconds part\r\n"))
			return
		}

		autoID := generateAutoID(stream, &millis)
		entryID = autoID
		args[2] = autoID
	} else if entryID == "*" {
		autoID := generateAutoID(stream, nil)
		entryID = autoID
		args[2] = autoID
	}

	if exists && len(stream.Entries) > 0 {
		lastEntryID := stream.Entries[len(stream.Entries)-1].ID
		if err := validateID(entryID, lastEntryID); err != nil {
			conn.Write([]byte(fmt.Sprintf("-%s\r\n", err.Error())))
			return
		}
	} else {
		if err := validateID(entryID, ""); err != nil {
			conn.Write([]byte(fmt.Sprintf("-%s\r\n", err.Error())))
			return
		}
	}

	stream.Entries = append(stream.Entries, core.StreamEntry{
		ID:     entryID,
		Fields: fieldMap,
	})

	store[streamKey] = core.StoreEntry{
		Type: "stream",
		Data: stream,
	}

	// Notify waiting clients
	waitingClientsMu.Lock()
	clients := waitingClients[streamKey]
	newClients := make([]*waitingClient, 0, len(clients))

	for _, client := range clients {
		startID, ok := client.streams[streamKey]
		if !ok {
			continue
		}

		compare, err := compareIDs(entryID, startID)
		if err != nil || compare <= 0 {
			newClients = append(newClients, client)
			continue
		}

		// Collect entries for all streams the client is waiting on
		responseEntries := make(map[string][]core.StreamEntry)
		for sKey, sStartID := range client.streams {
			entry, exists := store[sKey]
			if !exists || entry.Type != "stream" {
				continue
			}

			stream, ok := entry.Data.(core.Stream)
			if !ok {
				continue
			}

			var entries []core.StreamEntry
			for _, e := range stream.Entries {
				if compare, _ := compareIDs(e.ID, sStartID); compare > 0 {
					entries = append(entries, e)
				}
			}

			if len(entries) > 0 {
				responseEntries[sKey] = entries
			}
		}

		if len(responseEntries) > 0 {
			select {
			case client.responseCh <- xreadResponse{entries: responseEntries}:
				// Client notified
			default:
				// Could not notify, client may have timed out
			}
		} else {
			newClients = append(newClients, client)
		}
	}

	waitingClients[streamKey] = newClients
	waitingClientsMu.Unlock()

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

	sequenceNumber := int64(0)
	if parts[1] != "*" {
		seq, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid sequenceNumber")
		}
		sequenceNumber = seq
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
		if newMillis == 0 && newSeq == 1 {
			return nil
		}
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

func generateAutoID(stream core.Stream, explicitMillis *int64) string {
	var millis int64
	if explicitMillis != nil {
		millis = *explicitMillis
	} else {
		millis = time.Now().UnixMilli()
	}

	var seq int64 = 0

	if len(stream.Entries) > 0 {
		lastEntryID := stream.Entries[len(stream.Entries)-1].ID
		lastMillis, lastSeq, err := parseID(lastEntryID)
		if err == nil && lastMillis == millis {
			seq = lastSeq + 1
		}
	} else if millis == 0 {
		seq = 1
	}

	return fmt.Sprintf("%d-%d", millis, seq)
}
