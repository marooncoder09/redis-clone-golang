package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports in stage 1 (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {
	fmt.Println("Starting server on port 6379...")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Connection closed:", err.Error())
			return
		}
		msg = strings.TrimSpace(msg)

		argCount, err := strconv.Atoi(msg[1:])
		if err != nil || argCount < 1 {
			conn.Write([]byte("-ERR invalid argument count\r\n"))
			continue
		}

		args := make([]string, 0, argCount)

		for i := 0; i < argCount; i++ {
			_, err := reader.ReadString('\n')
			if err != nil {
				conn.Write([]byte("-ERR unexpected end of input\r\n"))
				return
			}

			arg, err := reader.ReadString('\n')
			if err != nil {
				conn.Write([]byte("-ERR unexpected end of input\r\n"))
				return
			}
			args = append(args, strings.TrimSpace(arg))
		}

		command := strings.ToUpper(args[0])

		switch command {
		case "PING":
			conn.Write([]byte("+PONG\r\n"))
		case "ECHO":
			if len(args) > 1 {
				response := args[1]
				conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(response), response)))
			} else {
				conn.Write([]byte("$0\r\n\r\n"))
			}
		default:
			conn.Write([]byte("-ERR unknown command\r\n"))
		}
	}
}
