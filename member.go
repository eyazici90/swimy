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

func (ms *Membership) setAlives(members ...*Member) {
	ms.membersMu.Lock()
	defer ms.membersMu.Unlock()

	now := time.Now().UTC()
	for _, m := range members {
		m.state = alive
		m.since = now
	}
}

func (ms *Membership) discovered(members ...*Member) {
	ms.membersMu.Lock()
	for _, m := range members {
		ms.others = append(ms.others, m)
	}
	ms.membersMu.Unlock()
}
