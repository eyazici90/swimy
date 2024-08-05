package swim

import "sync/atomic"

type Metrics struct {
	ActiveMembers        uint32
	SentNum, ReceivedNum uint32
}

type observation struct {
	metrics         Metrics
	onJoinCallback  func(m *Member)
	onLeaveCallback func(m *Member)
}

func (o *observation) onJoin(m *Member) {
	o.onJoinCallback(m)
	atomic.AddUint32(&o.metrics.ActiveMembers, 1)
}

func (o *observation) onLeave(m *Member) {
	o.onLeaveCallback(m)
	atomic.AddUint32(&o.metrics.ActiveMembers, -1)
}

func (o *observation) pinged() {
	atomic.AddUint32(&o.metrics.SentNum, 1)
}

func (o *observation) received() {
	atomic.AddUint32(&o.metrics.ReceivedNum, 1)
}
