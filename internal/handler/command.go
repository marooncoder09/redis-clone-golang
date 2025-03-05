package handler

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"

	parser "github.com/codecrafters-io/redis-starter-go/internal/parser"
)

var (
	store = make(map[string]string)
	mu    sync.RWMutex
)

func HandleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		args, err := parser.ParseRequest(reader)
		if err != nil {
			fmt.Println("Error parsing request:", err)
			return
		}
		if len(args) == 0 {
			continue
		}
		processCommand(conn, args)
	}
}

func processCommand(conn net.Conn, args []string) {
	command := strings.ToUpper(args[0])
	fmt.Println("Processing command:", command)
	switch command {
	case "PING":
		handlePing(conn)
	case "ECHO":
		handleEcho(conn, args)
	case "SET":
		handleSet(conn, args)
	case "GET":
		handleGet(conn, args)
	default:
		conn.Write([]byte("-ERR unknown command\r\n"))
	}
}

func handlePing(conn net.Conn) {
	conn.Write([]byte("+PONG\r\n"))
	return
}

func handleEcho(conn net.Conn, args []string) {
	if len(args) > 1 {
		response := args[1]
		conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(response), response)))
	} else {
		conn.Write([]byte("$0\r\n\r\n"))
	}
}

func handleSet(conn net.Conn, args []string) {
	fmt.Println("Setting key:", args[1], "to value:", args[2])
	if len(args) < 3 {
		conn.Write([]byte("-ERR wrong number of arguments for 'SET' command\r\n"))
		return
	}
	mu.Lock()
	store[args[1]] = args[2]
	mu.Unlock()
	conn.Write([]byte("+OK\r\n"))
}

func handleGet(conn net.Conn, args []string) {
	if len(args) < 2 {
		conn.Write([]byte("-ERR wrong number of arguments for 'GET' command\r\n"))
		return
	}
	mu.RLock()
	value, exists := store[args[1]]
	mu.RUnlock()
	if !exists {
		conn.Write([]byte("$-1\r\n"))
	} else {
		conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)))
	}
}
