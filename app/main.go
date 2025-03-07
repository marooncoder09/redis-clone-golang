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
	flag.Parse()

	commands.SetConfig("dir", *dir)
	commands.SetConfig("dbfilename", *dbfilename)

	rdbPath := filepath.Join(*dir, *dbfilename)
	entries, err := parser.ParseRDB(rdbPath)
	if err != nil {
		log.Fatal("Error loading RDB file:", err)
	}

	for key, value := range entries {
		commands.SetKeyEntry(key, value)
	}

	fmt.Println("Starting server on port 6379...")
	err = server.Start("0.0.0.0:6379")
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
