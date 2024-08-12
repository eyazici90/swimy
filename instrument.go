package swim

import (
	"log"
	"net"
	"sync/atomic"
)

type Metrics struct {
	ActiveMembers        uint32
	SentNum, ReceivedNum uint32
}

type observation struct {
	me              net.Addr
	metrics         Metrics
	onJoinCallback  func(m *Member)
	onLeaveCallback func(m *Member)
}

func (o *observation) onJoin(m *Member) {
	o.onJoinCallback(m)
	atomic.AddUint32(&o.metrics.ActiveMembers, 1)
	log.Printf("me: %s, someone joined addr: %s", o.me, m.addr)
}

func (o *observation) onLeave(m *Member) {
	o.onLeaveCallback(m)
	atomic.AddUint32(&o.metrics.ActiveMembers, ^uint32(0))
	log.Printf("me: %s, someone left addr", o.me)
}

func (o *observation) onStop() {
	log.Printf("stopped addr: %s", o.me) // move this to observer
}

func (o *observation) pinged() {
	atomic.AddUint32(&o.metrics.SentNum, 1)
}

func (o *observation) received(msg, addr string) {
	atomic.AddUint32(&o.metrics.ReceivedNum, 1)
	log.Printf("me: %s, received %s from: %s", o.me, msg, addr)
}
