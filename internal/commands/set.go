package commands

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/replication"
)

func HandleSet(conn net.Conn, args []string) {
	if len(args) < 3 {
		conn.Write([]byte("-ERR wrong number of arguments for 'SET' command\r\n"))
		return
	}

	key := args[1]
	value := args[2]
	var ttl int64 = 0

	if len(args) >= 5 && strings.ToUpper(args[3]) == "PX" {
		parsedTTL, err := strconv.ParseInt(args[4], 10, 64)
		if err != nil || parsedTTL <= 0 {
			conn.Write([]byte("-ERR PX must be a positive integer\r\n"))
			return
		}
		ttl = parsedTTL
	}

	SetKey(key, value, ttl)

	conn.Write([]byte("+OK\r\n"))
	fmt.Println("Processed SET:", key, "->", value, "PX:", ttl)

	replication.PropagateCommand("SET", args[1:])
}
