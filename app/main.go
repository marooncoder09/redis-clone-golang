package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

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
