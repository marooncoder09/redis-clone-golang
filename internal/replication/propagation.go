package replication

import (
	"fmt"
	"net"
	"sync"
)

var (
	mu       sync.RWMutex
	replicas []net.Conn // List of connected replicas
)

func AddReplica(conn net.Conn) {
	mu.Lock()
	defer mu.Unlock()
	replicas = append(replicas, conn)
}

func PropagateCommand(command string, args []string) {
	mu.RLock()
	defer mu.RUnlock()

	if len(replicas) == 0 {
		fmt.Println("[Master] No replicas to propagate to.")
		return
	}

	respCommand := encodeCommandRESP(command, args)

	fmt.Println("[Master] Propagating command:", command, args)

	for _, replica := range replicas {
		_, err := replica.Write([]byte(respCommand))
		if err != nil {
			fmt.Println("[Master] Failed to propagate to replica:", err)
		}
	}
}

func encodeCommandRESP(command string, args []string) string {
	resp := fmt.Sprintf("*%d\r\n$%d\r\n%s\r\n", len(args)+1, len(command), command)
	for _, arg := range args {
		resp += fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg)
	}
	return resp
}
