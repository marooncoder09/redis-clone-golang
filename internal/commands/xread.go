package commands

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/models/core"
)

func HandleXread(conn net.Conn, args []string) {
	var blockTimeoutMillis int64 = -1
	streamsArgsIndex := 1

	if len(args) >= 2 && strings.ToUpper(args[1]) == "BLOCK" {
		if len(args) < 5 {
			conn.Write([]byte("-ERR wrong number of arguments for 'XREAD' command\r\n"))
			return
		}

		timeoutStr := args[2]
		timeout, err := strconv.ParseInt(timeoutStr, 10, 64)
		if err != nil {
			conn.Write([]byte("-ERR invalid timeout value\r\n"))
			return
		}
		blockTimeoutMillis = timeout

		if strings.ToUpper(args[3]) != "STREAMS" {
			conn.Write([]byte("-ERR expected STREAMS keyword\r\n"))
			return
		}
		streamsArgsIndex = 4
	} else {
		if len(args) < 2 || strings.ToUpper(args[1]) != "STREAMS" {
			conn.Write([]byte("-ERR wrong number of arguments for 'XREAD' command\r\n"))
			return
		}
		streamsArgsIndex = 2
	}

	remainingArgs := args[streamsArgsIndex:]
	if len(remainingArgs)%2 != 0 || len(remainingArgs) == 0 {
		conn.Write([]byte("-ERR wrong number of arguments for streams and IDs\r\n"))
		return
	}

	numStreams := len(remainingArgs) / 2
	streamKeys := remainingArgs[:numStreams]
	startIDs := remainingArgs[numStreams:]

	mu.RLock()
	responseEntries := make(map[string][]core.StreamEntry)
	entriesAvailable := false
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

		if startID == "$" {

			if len(stream.Entries) > 0 {
				startID = stream.Entries[len(stream.Entries)-1].ID
			} else {

				startID = "0-0"
			}
		}

		for _, e := range stream.Entries {
			compare, err := compareIDs(e.ID, startID)
			if err == nil && compare > 0 {
				entries = append(entries, e)
			}
		}

		if len(entries) > 0 {
			responseEntries[streamKey] = entries
			entriesAvailable = true
		}
	}
	mu.RUnlock()

	if entriesAvailable || blockTimeoutMillis < 0 {
		sendXreadResponse(conn, responseEntries)
		return
	}

	// Blocking mode setup
	wc := &waitingClient{
		streams:    make(map[string]string),
		responseCh: make(chan xreadResponse, 1),
		deadline:   time.Now().Add(time.Duration(blockTimeoutMillis) * time.Millisecond),
	}

	for i, streamKey := range streamKeys {
		startID := startIDs[i]
		if startID == "$" {
			entry, exists := store[streamKey]
			if exists && entry.Type == "stream" {
				stream, ok := entry.Data.(core.Stream)
				if ok && len(stream.Entries) > 0 {
					startID = stream.Entries[len(stream.Entries)-1].ID
				} else {
					startID = "0-0"
				}
			}
		}
		wc.streams[streamKey] = startID
	}

	// Add to waiting clients
	waitingClientsMu.Lock()
	for streamKey := range wc.streams {
		waitingClients[streamKey] = append(waitingClients[streamKey], wc)
	}
	waitingClientsMu.Unlock()

	// Wait for response or timeout
	if blockTimeoutMillis == 0 {
		// Blocking indefinitely
		res := <-wc.responseCh
		sendXreadResponse(conn, res.entries)
	} else {
		// Blocking with timeout
		select {
		case res := <-wc.responseCh:
			sendXreadResponse(conn, res.entries)
		case <-time.After(time.Until(wc.deadline)):
			mu.RLock()
			timeoutResponse := make(map[string][]core.StreamEntry)
			for streamKey, startID := range wc.streams {
				entry, exists := store[streamKey]
				if !exists || entry.Type != "stream" {
					continue
				}

				stream, ok := entry.Data.(core.Stream)
				if !ok {
					continue
				}

				var entries []core.StreamEntry
				for _, e := range stream.Entries {
					compare, _ := compareIDs(e.ID, startID)
					if compare > 0 {
						entries = append(entries, e)
					}
				}

				if len(entries) > 0 {
					timeoutResponse[streamKey] = entries
				}
			}
			mu.RUnlock()

			if len(timeoutResponse) > 0 {
				sendXreadResponse(conn, timeoutResponse)
			} else {
				conn.Write([]byte("$-1\r\n"))
			}
		}
	}

	// Cleanup
	waitingClientsMu.Lock()
	for streamKey := range wc.streams {
		for i, client := range waitingClients[streamKey] {
			if client == wc {
				waitingClients[streamKey] = append(waitingClients[streamKey][:i], waitingClients[streamKey][i+1:]...)
				break
			}
		}
	}
	waitingClientsMu.Unlock()
}

func sendXreadResponse(conn net.Conn, entries map[string][]core.StreamEntry) {
	if len(entries) == 0 {
		conn.Write([]byte("*-1\r\n"))
		return
	}

	var resp strings.Builder
	resp.WriteString(fmt.Sprintf("*%d\r\n", len(entries)))

	for streamKey, streamEntries := range entries {
		resp.WriteString(fmt.Sprintf("*2\r\n$%d\r\n%s\r\n", len(streamKey), streamKey))
		resp.WriteString(fmt.Sprintf("*%d\r\n", len(streamEntries)))

		for _, entry := range streamEntries {
			resp.WriteString(fmt.Sprintf("*2\r\n$%d\r\n%s\r\n", len(entry.ID), entry.ID))

			var fields []string
			for k, v := range entry.Fields {
				fields = append(fields, k, v)
			}

			resp.WriteString(fmt.Sprintf("*%d\r\n", len(fields)))
			for _, f := range fields {
				resp.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(f), f))
			}
		}
	}

	conn.Write([]byte(resp.String()))
}

// func encodeXreadResponse(response []interface{}) string {
// 	if len(response) == 0 {
// 		return "*0\r\n"
// 	}

// 	var resp strings.Builder
// 	resp.WriteString(fmt.Sprintf("*%d\r\n", len(response)))

// 	for _, streamResponse := range response {
// 		streamData := streamResponse.([]interface{})
// 		streamKey := streamData[0].(string)
// 		entries := streamData[1].([]core.StreamEntry)

// 		resp.WriteString(fmt.Sprintf("*2\r\n$%d\r\n%s\r\n", len(streamKey), streamKey))
// 		resp.WriteString(fmt.Sprintf("*%d\r\n", len(entries)))

// 		for _, entry := range entries {
// 			resp.WriteString(fmt.Sprintf("*2\r\n$%d\r\n%s\r\n", len(entry.ID), entry.ID))

// 			var fields []string
// 			for key, value := range entry.Fields {
// 				fields = append(fields, key, value)
// 			}

// 			resp.WriteString(fmt.Sprintf("*%d\r\n", len(fields)))
// 			for _, field := range fields {
// 				resp.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(field), field))
// 			}
// 		}
// 	}

// 	return resp.String()
// }
