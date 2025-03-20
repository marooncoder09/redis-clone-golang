package commands

import (
	"net"
	"sync"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/models/core"
)

var (
	store   = make(map[string]core.StoreEntry)
	mu      sync.RWMutex
	configs = map[string]string{
		"dir":        "/tmp", // setting this to tmp for now.
		"dbfilename": "dump.rdb",
	}
	replicas []net.Conn

	waitingClientsMu sync.Mutex
	waitingClients   = make(map[string][]*waitingClient)
)

type waitingClient struct {
	streams    map[string]string
	responseCh chan xreadResponse
	deadline   time.Time
}

type xreadResponse struct {
	entries map[string][]core.StreamEntry
}

func SetKey(key, value string, ttl int64) {
	mu.Lock()
	defer mu.Unlock()

	entry := core.StoreEntry{Data: value, Type: "string"}
	if ttl > 0 {
		entry.ExpiresAt = time.Now().UnixMilli() + ttl // Expiry in milliseconds, should i make it in seconds?
	}
	store[key] = entry
}

func GetKey(key string) (core.StoreEntry, bool) {
	mu.RLock()
	entry, exists := store[key]
	mu.RUnlock()

	if !exists {
		return core.StoreEntry{}, false
	}

	if entry.ExpiresAt > 0 && time.Now().UnixMilli() > entry.ExpiresAt {
		mu.Lock()
		delete(store, key)
		mu.Unlock()
		return core.StoreEntry{}, false

		// TODO: Handle the expiration part in a better way.
		/*
			the current impl will delete the keys from the store when the user try to access it
			however if the user does not try to access the key, the key will not be deleted from the store
			for this imo what we can do in the future is to have a cron job that will delete the exp. keys form the store.

			ISSUES IN CRON IMPL:
			1. somehow we have to make sure that the cron job is not breaking the main thread
			2. we make to make the crons thead safe
			3. we have to make sure that the cron job is not using too much CPU
			4. and also make sure that in case of too much load on the server the cron job is not started,
			   dynamically adjust the cron for the memory usage
			5. make sure to delete the keys in batch and not 1 by 1 to avoide the overhead
		*/
	}

	return entry, true
}

func SetConfig(key, value string) {
	mu.Lock()
	defer mu.Unlock()
	configs[key] = value
}

func GetConfig(key string) (string, bool) {
	mu.RLock()
	defer mu.RUnlock()
	value, exists := configs[key]
	return value, exists
}

func SetKeyEntry(key string, entry core.StoreEntry) {
	mu.Lock()
	defer mu.Unlock()
	store[key] = entry
}

func ClearStore() {
	mu.Lock()
	defer mu.Unlock()
	store = make(map[string]core.StoreEntry)
}

func GetEntry(key string) (core.StoreEntry, bool) {
	mu.RLock()
	entry, exists := store[key]
	mu.RUnlock()

	if !exists {
		return core.StoreEntry{}, false
	}

	if entry.ExpiresAt > 0 && time.Now().UnixMilli() > entry.ExpiresAt {
		mu.Lock()
		delete(store, key)
		mu.Unlock()
		return core.StoreEntry{}, false
	}

	return entry, true
}
