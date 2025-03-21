package core

import (
	"net"
	"sync"
)

type ClientState struct {
	InTransaction bool
}

var (
	ClientStates = make(map[net.Conn]*ClientState)
	ClientMu     sync.Mutex
)
