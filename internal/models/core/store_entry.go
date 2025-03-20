package core

type StoreEntry struct {
	Data      interface{}
	ExpiresAt int64
	Type      string
}
