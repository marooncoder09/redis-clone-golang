package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"path/filepath"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/commands"
	"github.com/codecrafters-io/redis-starter-go/internal/parser"
	"github.com/codecrafters-io/redis-starter-go/internal/server"
)

func main() {
	dir := flag.String("dir", "/tmp/redis-data", "directory for RDB files")
	dbfilename := flag.String("dbfilename", "dump.rdb", "RDB file name")
	port := flag.String("port", "6379", "port to run the server on")
	replicaof := flag.String("replicaof", "", "Master host and port (for replica mode)")

	flag.Parse()

	commands.SetConfig("dir", *dir)
	commands.SetConfig("dbfilename", *dbfilename)

	if *replicaof != "" { // here it is master by default, but if the flag replicaof is set then it is a slave
		commands.SetConfig("role", "slave")
		go connectToMaster(*replicaof)

	} else {
		commands.SetConfig("role", "master")
	}

	commands.SetConfig("master_replid", "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb") // hardcoded for now as per the challenge
	commands.SetConfig("master_repl_offset", "0")

	rdbPath := filepath.Join(*dir, *dbfilename)
	entries, err := parser.ParseRDB(rdbPath)
	if err != nil {
		log.Fatal("Error loading RDB file:", err)
	}

	for key, value := range entries {
		commands.SetKeyEntry(key, value)
	}

	fmt.Println("Starting server on port", *port, "...")
	err = server.Start("0.0.0.0:" + *port)
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func connectToMaster(masterAddress string) {
	parts := strings.Split(masterAddress, " ")
	if len(parts) != 2 {
		log.Println("Invalid --replicaof format. Expected: '<MASTER_HOST> <MASTER_PORT>'")
		return
	}

	host, port := parts[0], parts[1]
	address := fmt.Sprintf("%s:%s", host, port)

	fmt.Println("Connecting to master at", address)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Println("Failed to connect to master:", err)
		return
	}

	fmt.Println("Connected to master. Sending PING...")

	pingMessage := "*1\r\n$4\r\nPING\r\n"
	_, err = conn.Write([]byte(pingMessage))
	if err != nil {
		log.Println("Failed to send PING to master:", err)
		conn.Close()
		return
	}

	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Error reading response from master:", err)
		conn.Close()
		return
	}

	fmt.Println("Received response from master:", strings.TrimSpace(response))

	for {
		time.Sleep(time.Second)
	}
}
