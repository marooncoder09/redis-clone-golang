package core

type Stream struct {
	Entries []StreamEntry
}

type StreamEntry struct {
	ID     string
	Fields map[string]string
}
