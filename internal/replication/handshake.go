package replication

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

// performHandshake handles PING and REPLCONF handshake with the master
func performHandshake(conn net.Conn, replicaPort string) {
	reader := bufio.NewReader(conn)

	// step 1: send PING
	pingMessage := "*1\r\n$4\r\nPING\r\n"
	fmt.Println("Sending PING to master...")
	_, err := conn.Write([]byte(pingMessage))
	if err != nil {
		log.Println("Failed to send PING to master:", err)
		conn.Close()
		return
	}

	// step 2: read PONG response
	response, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Error reading PONG response:", err)
		conn.Close()
		return
	}
	fmt.Println("Received from master:", strings.TrimSpace(response))

	// step 3: send REPLCONF listening-port <PORT>
	replconfListening := fmt.Sprintf("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$%d\r\n%s\r\n", len(replicaPort), replicaPort)
	fmt.Println("Sending REPLCONF listening-port", replicaPort)
	_, err = conn.Write([]byte(replconfListening))
	if err != nil {
		log.Println("Failed to send REPLCONF listening-port:", err)
		conn.Close()
		return
	}

	// step 4: read OK response
	response, err = reader.ReadString('\n')
	if err != nil {
		log.Println("Error reading response to REPLCONF listening-port:", err)
		conn.Close()
		return
	}
	fmt.Println("Received from master:", strings.TrimSpace(response))

	// step 5: send REPLCONF capa psync2
	replconfCapa := "*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n"
	fmt.Println("Sending REPLCONF capa psync2")
	_, err = conn.Write([]byte(replconfCapa))
	if err != nil {
		log.Println("Failed to send REPLCONF capa psync2:", err)
		conn.Close()
		return
	}

	// step 6: read OK response
	response, err = reader.ReadString('\n')
	if err != nil {
		log.Println("Error reading response to REPLCONF capa:", err)
		conn.Close()
		return
	}
	fmt.Println("Received from master:", strings.TrimSpace(response))

	for {
		time.Sleep(time.Second)
	}
}
