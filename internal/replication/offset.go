package replication

import "sync"

var (
	globalOffset int64
	offsetMu     sync.RWMutex
)

func GetOffset() int64 {
	offsetMu.RLock()
	defer offsetMu.RUnlock()
	return globalOffset
}

func AddToOffset(n int64) {
	offsetMu.Lock()
	defer offsetMu.Unlock()
	globalOffset += n
}

func SetOffset(val int64) {
	offsetMu.Lock()
	defer offsetMu.Unlock()
	globalOffset = val
}
