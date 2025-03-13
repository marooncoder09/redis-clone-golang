package commands

import (
	"fmt"
	"net"
)

func HandleKeys(conn net.Conn, args []string) {
	mu.RLock()
	defer mu.RUnlock()

	keysList := make([]string, 0, len(store)) // Preallocating the slice for better performance

	for key := range store {
		keysList = append(keysList, key)
	}

	// RESP format: "*<count>\r\n$<length>\r\n<key>\r\n"
	resp := fmt.Sprintf("*%d\r\n", len(keysList))
	for _, key := range keysList {
		resp += fmt.Sprintf("$%d\r\n%s\r\n", len(key), key)
	}

	conn.Write([]byte(resp))
}

/* TODO:
1. Here we have to impl. that when getting the key if the key has
expired we have to delete that particular key as well and return the
remaining (similar to what we did in the GET), besides this also impl
the CRON job to delete the expired keys
*/
