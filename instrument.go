package swim

import (
	"context"
	"io"
	"log"
	"log/slog"
	"net"
	"os"
	"sync/atomic"
)

type Metrics struct {
	ActiveMembers        uint32
	SentNum, ReceivedNum uint32
}

type defaultObserver struct {
	me              net.Addr
	metrics         Metrics
	onJoinCallback  func(m net.Addr)
	onLeaveCallback func(m net.Addr)
}

func (o *defaultObserver) onJoin(addr net.Addr) {
	o.onJoinCallback(addr)
	atomic.AddUint32(&o.metrics.ActiveMembers, 1)
	log.Printf("me: %s, someone joined addr: %s", o.me, addr)
}

func (o *defaultObserver) onLeave(addr net.Addr) {
	o.onLeaveCallback(addr)
	atomic.AddUint32(&o.metrics.ActiveMembers, ^uint32(0))
	log.Printf("me: %s, someone left addr: %s", o.me, addr)
}

func (o *defaultObserver) onStop() {
	log.Printf("stopped addr: %s", o.me) // move this to observer
}

func (o *defaultObserver) onSilentErr(ctx context.Context, err error) {
	slog.Log(ctx,
		slog.LevelError,
		err.Error(),
	)
}

func (o *defaultObserver) pinged() {
	atomic.AddUint32(&o.metrics.SentNum, 1)
}

func (o *defaultObserver) received(msg, addr string) {
	atomic.AddUint32(&o.metrics.ReceivedNum, 1)
	log.Printf("me: %s, received %s from: %s", o.me, msg, addr)
}

func (ms *Membership) useDefaultObserver() {
	ms.observer = &defaultObserver{
		me:              ms.me.Addr(),
		onJoinCallback:  ms.cfg.OnJoin,
		onLeaveCallback: ms.cfg.OnLeave,
	}
	setUpSlog(os.Stdout)
}

func setUpSlog(wr io.Writer) {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	h := slog.NewTextHandler(wr, opts)
	sl := slog.New(h)

	slog.SetDefault(sl)
}
