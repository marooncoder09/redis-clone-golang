package replication

import (
	"fmt"
	"log"
	"net"
	"sync"
)

var (
	mu             sync.RWMutex
	replicas       []*ReplicaConn
	replicaOffsets = make(map[net.Conn]int64)
)

type ReplicaConn struct {
	conn    net.Conn
	writeCh chan []byte
}

func NewReplicaConn(conn net.Conn) *ReplicaConn {
	rc := &ReplicaConn{
		conn:    conn,
		writeCh: make(chan []byte, 100),
	}

	go func() {
		for data := range rc.writeCh {
			fmt.Printf("[DEBUG-WRITER] Actually writing to %v: %q\n", conn.RemoteAddr(), data)

			_, err := conn.Write(data)
			if err != nil {

				fmt.Println("[Master] Failed to write to replica:", err)
			}
		}
	}()

	return rc
}

func (r *ReplicaConn) Enqueue(data []byte) {
	r.writeCh <- data
}

func AddReplica(conn net.Conn) {
	mu.Lock()
	defer mu.Unlock()

	for _, r := range replicas {
		if r.conn == conn {
			return
		}
	}

	rc := NewReplicaConn(conn)
	replicas = append(replicas, rc)
	replicaOffsets[conn] = 0
}

func PropagateCommand(command string, args []string) {
	mu.RLock()
	defer mu.RUnlock()

	if len(replicas) == 0 {
		fmt.Println("[Master] No replicas to propagate to.")
		return
	}

	respCommand := encodeCommandRESP(command, args)
	byteCount := len(respCommand)

	AddToOffset(int64(byteCount))

	fmt.Println("[Master] Propagating command:", command, args)
	fmt.Printf("[DEBUG] Propagating: %s / %v (bytes: %d)\n", command, args, byteCount)

	for _, rc := range replicas {
		rc.Enqueue([]byte(respCommand))
	}
}

func encodeCommandRESP(command string, args []string) string {
	resp := fmt.Sprintf("*%d\r\n$%d\r\n%s\r\n", len(args)+1, len(command), command)
	for _, arg := range args {
		resp += fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg)
	}
	return resp
}

func GetReplicaCount() int {
	mu.RLock()
	defer mu.RUnlock()
	return len(replicas)
}

func RequestAckFromReplicas() {
	cmd := "*3\r\n$8\r\nREPLCONF\r\n$6\r\nGETACK\r\n$1\r\n*\r\n"
	fmt.Println("[DEBUG] Sending GETACK to all replicas")
	mu.RLock()
	defer mu.RUnlock()
	for _, rc := range replicas {
		fmt.Printf("[DEBUG] Sending GETACK to %v\n", rc.conn.RemoteAddr())
		rc.Enqueue([]byte(cmd))
	}
}

func SetReplicaOffset(conn net.Conn, offset int64) {
	mu.Lock()
	defer mu.Unlock()

	oldOffset, exists := replicaOffsets[conn]
	fmt.Printf("[DEBUG] SetReplicaOffset: conn=%p oldOffset=%d newOffset=%d exists=%v\n",
		conn, oldOffset, offset, exists)

	replicaOffsets[conn] = offset
}

func CountReplicasAtOrAboveOffset(offset int64) int {
	mu.RLock()
	defer mu.RUnlock()
	count := 0
	for c, off := range replicaOffsets {
		fmt.Printf("[DEBUG] Checking conn=%p offset=%d vs desired=%d\n", c, off, offset)
		if off >= offset {
			count++
		}
	}
	fmt.Printf("[DEBUG] CountReplicasAtOrAboveOffset => %d\n", count)
	return count
}

func EnqueueResponseForReplica(conn net.Conn, msg string) {
	if _, err := conn.Write([]byte(msg)); err != nil {
		log.Println("Failed to send ACK:", err)
	}
}
