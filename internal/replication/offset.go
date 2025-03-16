package replication

import (
	"fmt"
	"sync"
)

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
	fmt.Printf("[OFFSET] Adding %d to offset (was %d)\n", n, globalOffset)

	globalOffset += n
}

func SetOffset(val int64) {
	offsetMu.Lock()
	defer offsetMu.Unlock()
	globalOffset = val
}
