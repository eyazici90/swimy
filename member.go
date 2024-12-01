package swimy

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

func newAliveMember(addr net.Addr) Member {
	return Member{
		addr:  addr,
		state: alive,
		since: time.Now().UTC(),
	}
}

func (m Member) Addr() net.Addr {
	return m.addr
}

func (ms *Membership) alives(excludes ...net.Addr) []Member {
	ms.membersMu.RLock()
	defer ms.membersMu.RUnlock()

	var res []Member
	for _, m := range ms.others {
		if slices.Contains(excludes, m.Addr()) {
			continue
		}
		if m.state == alive {
			res = append(res, m)
		}
	}
	return res
}

func (ms *Membership) setAlives(members ...Member) {
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
	for _, addr := range addrs {
		if m, ok := ms.others[addr]; ok {
			m.state = state
			m.since = now
			ms.others[addr] = m
		}
	}
}

func (ms *Membership) becomeMembers(members ...Member) {
	ms.membersMu.Lock()
	for _, m := range members {
		ms.others[m.Addr()] = m
	}
	ms.membersMu.Unlock()
}
