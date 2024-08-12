package swim

import (
	"net"
	"slices"
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

func (ms *Membership) isAware(addr net.Addr) bool {
	ms.membersMu.RLock()
	defer ms.membersMu.RUnlock()

	for _, m := range ms.others {
		if m.Addr() == addr {
			return true
		}
	}
	return false
}

func (ms *Membership) alives() []*Member {
	ms.membersMu.RLock()
	defer ms.membersMu.RUnlock()

	var res []*Member
	for _, m := range ms.others {
		if m.state == alive {
			res = append(res, m)
		}
	}
	return res
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

func (ms *Membership) setState(state memberState, addrs ...net.Addr) {
	ms.membersMu.Lock()
	defer ms.membersMu.Unlock()

	now := time.Now().UTC()
	for _, m := range ms.others {
		if slices.Contains(addrs, m.Addr()) {
			m.state = state
			m.since = now
		}
	}
}

func (ms *Membership) becomeMembers(members ...*Member) {
	ms.membersMu.Lock()
	ms.others = append(ms.others, members...)
	ms.membersMu.Unlock()
}
