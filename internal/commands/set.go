package commands

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/models/core"
	"github.com/codecrafters-io/redis-starter-go/internal/replication"
)

var (
	db = make(map[string]string)
)

func HandleSet(conn net.Conn, args []string, isReplica bool) {
	if len(args) < 3 {
		if conn != nil {
			conn.Write([]byte("-ERR wrong number of arguments for 'SET' command\r\n"))
		}
		return
	}

	key := args[1]
	value := args[2]
	var ttl int64 = 0

	if len(args) >= 5 && strings.ToUpper(args[3]) == "PX" {
		parsedTTL, err := strconv.ParseInt(args[4], 10, 64)
		if err != nil || parsedTTL <= 0 {
			if conn != nil {
				conn.Write([]byte("-ERR PX must be a positive integer\r\n"))
			}
			return
		}
		ttl = parsedTTL
	}

	entry := core.StoreEntry{Value: value}
	if ttl > 0 {
		entry.ExpiresAt = time.Now().UnixMilli() + ttl
	}

	if isReplica {
		SetKeyEntry(key, entry)
		fmt.Println("[Replica] Stored key:", key, "Value:", value, "TTL:", ttl)
	} else {
		SetKey(key, value, ttl)

		if conn != nil {
			conn.Write([]byte("+OK\r\n"))
		}

		fmt.Println("[Master] Processed SET:", key, "->", value, "TTL:", ttl)

		replication.PropagateCommand("SET", args[1:])
	}
}
