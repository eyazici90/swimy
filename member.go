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

type member struct {
	addr  net.Addr
	state memberState
	since time.Time
}
