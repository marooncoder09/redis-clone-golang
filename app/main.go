package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/codecrafters-io/redis-starter-go/internal/commands"
	"github.com/codecrafters-io/redis-starter-go/internal/server"
)

func main() {
	dir := flag.String("dir", "/tmp/redis-data", "directory for RDB files")
	dbfilename := flag.String("dbfilename", "dump.rdb", "RDB file name")
	flag.Parse()

	// initial config values
	commands.SetConfig("dir", *dir)
	commands.SetConfig("dbfilename", *dbfilename)

	fmt.Println("Starting server on port 6379...")
	err := server.Start("0.0.0.0:6379")
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
