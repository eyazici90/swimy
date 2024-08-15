package swim

import (
	"context"
	"io"
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

func (o *defaultObserver) onJoin(ctx context.Context, addr net.Addr) {
	o.onJoinCallback(addr)
	atomic.AddUint32(&o.metrics.ActiveMembers, 1)
	attr := slog.String("me", o.me.String())
	slog.LogAttrs(ctx, slog.LevelInfo, "someone joined", attr)
}

func (o *defaultObserver) onLeave(ctx context.Context, addr net.Addr) {
	o.onLeaveCallback(addr)
	atomic.AddUint32(&o.metrics.ActiveMembers, ^uint32(0))
	attr := slog.String("me", o.me.String())
	slog.LogAttrs(ctx, slog.LevelInfo, "someone left", attr)
}

func (o *defaultObserver) onStop() {
	attr := slog.String("me", o.me.String())
	slog.LogAttrs(nil, slog.LevelInfo, "stopped addr", attr)
}

func (o *defaultObserver) onSilentErr(ctx context.Context, err error) {
	attr := slog.String("me", o.me.String())
	slog.LogAttrs(ctx, slog.LevelError, err.Error(), attr)
}

func (o *defaultObserver) pinged() {
	atomic.AddUint32(&o.metrics.SentNum, 1)
}

func (o *defaultObserver) received(ctx context.Context, msg, senderAddr string) {
	atomic.AddUint32(&o.metrics.ReceivedNum, 1)
	attrs := []slog.Attr{
		slog.String("me", o.me.String()),
		slog.String("msg-type", msg),
		slog.String("sender", senderAddr),
	}
	slog.LogAttrs(ctx, slog.LevelInfo, "received", attrs...)
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
