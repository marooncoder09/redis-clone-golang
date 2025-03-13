package replication

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

func performHandshake(conn net.Conn, replicaPort string) {
	reader := bufio.NewReader(conn)

	pingMessage := "*1\r\n$4\r\nPING\r\n"
	fmt.Println("Sending PING to master...")
	_, err := conn.Write([]byte(pingMessage))
	if err != nil {
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

	replconfListening := fmt.Sprintf("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$%d\r\n%s\r\n", len(replicaPort), replicaPort)
	fmt.Println("Sending REPLCONF listening-port", replicaPort)
	_, err = conn.Write([]byte(replconfListening))
	if err != nil {
		log.Println("Failed to send REPLCONF listening-port:", err)
		conn.Close()
		return
	}

	response, err = reader.ReadString('\n')
	if err != nil {
		log.Println("Error reading response to REPLCONF listening-port:", err)
		conn.Close()
		return
	}
	fmt.Println("Received from master:", strings.TrimSpace(response))

	replconfCapa := "*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n" //this is the only capa that the master will accept
	/*
		what is capa?
		https://redis.io/docs/latest/replication/#replication-capabilities
	*/
	fmt.Println("Sending REPLCONF capa psync2")
	_, err = conn.Write([]byte(replconfCapa))
	if err != nil {
		log.Println("Failed to send REPLCONF capa psync2:", err)
		conn.Close()
		return
	}

	response, err = reader.ReadString('\n')
	if err != nil {
		log.Println("Error reading response to REPLCONF capa:", err)
		conn.Close()
		return
	}
	fmt.Println("Received from master:", strings.TrimSpace(response))

	psyncMessage := "*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n"
	fmt.Println("Sending PSYNC ? -1")
	_, err = conn.Write([]byte(psyncMessage))
	if err != nil {
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

	for {
		time.Sleep(time.Second)
	}
}
