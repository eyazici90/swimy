package swim

import (
	"net"
	"time"
)

type memberState int

const (
	unknown memberState = iota
	alive
	dead
	suspect
	left
)

type Member struct {
	addr  net.Addr
	state memberState
	since time.Time
}

func (m *Member) Addr() net.Addr {
	return m.addr
}
