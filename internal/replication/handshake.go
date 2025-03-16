package replication

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
)

func performHandshake(conn net.Conn, reader *bufio.Reader, replicaPort string) {

	pingMessage := "*1\r\n$4\r\nPING\r\n"
	fmt.Println("Sending PING to master...")
	if _, err := conn.Write([]byte(pingMessage)); err != nil {
		log.Println("Failed to send PING to master:", err)
		conn.Close()
		return
	}

	response, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Error reading PONG response:", err)
		conn.Close()
		return
	}
	fmt.Println("Received from master:", strings.TrimSpace(response))

	replconfListening := fmt.Sprintf(
		"*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$%d\r\n%s\r\n",
		len(replicaPort),
		replicaPort,
	)
	fmt.Println("Sending REPLCONF listening-port", replicaPort)
	if _, err := conn.Write([]byte(replconfListening)); err != nil {
		log.Println("Failed to send REPLCONF listening-port:", err)
		conn.Close()
		return
	}

	response, err = reader.ReadString('\n')
	if err != nil {
		log.Println("Error reading REPLCONF listening-port response:", err)
		conn.Close()
		return
	}
	fmt.Println("Received from master:", strings.TrimSpace(response))

	replconfCapa := "*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$3\r\neof\r\n"
	fmt.Println("Sending REPLCONF capa eof")
	if _, err := conn.Write([]byte(replconfCapa)); err != nil {
		log.Println("Failed to send REPLCONF capa command:", err)
		conn.Close()
		return
	}

	response, err = reader.ReadString('\n')
	if err != nil {
		log.Println("Error reading REPLCONF capa response:", err)
		conn.Close()
		return
	}
	fmt.Println("Received from master:", strings.TrimSpace(response))

	psyncMessage := "*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n"
	fmt.Println("Sending PSYNC ? -1")
	if _, err := conn.Write([]byte(psyncMessage)); err != nil {
		log.Println("Failed to send PSYNC:", err)
		conn.Close()
		return
	}

	response, err = reader.ReadString('\n')
	if err != nil {
		log.Println("Error reading response to PSYNC:", err)
		conn.Close()
		return
	}
	fmt.Println("Received from master:", strings.TrimSpace(response))

	bulkHeader, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Error reading RDB bulk header:", err)
		conn.Close()
		return
	}
	bulkHeader = strings.TrimSpace(bulkHeader)
	if !strings.HasPrefix(bulkHeader, "$") {
		log.Println("Invalid RDB bulk header:", bulkHeader)
		conn.Close()
		return
	}

	lengthStr := bulkHeader[1:]
	rdbLength, err := strconv.Atoi(lengthStr)
	if err != nil {
		log.Println("Error parsing RDB length:", err)
		conn.Close()
		return
	}

	rdbData := make([]byte, rdbLength)
	if _, err := io.ReadFull(reader, rdbData); err != nil {
		log.Println("Error reading RDB data:", err)
		conn.Close()
		return
	}

	SetOffset(0)
	fmt.Println("Handshake with master complete. Ready to receive commands.")
}
