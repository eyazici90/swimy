package swimy

import (
	"context"
	"math/rand/v2"
)

func (ms *Membership) suspect(ctx context.Context, member Member) (alive bool) {
	ms.setState(statusSuspect, member.Addr())
	req := indirectPingReq{sender: ms.me.Addr(), suspect: member.Addr()}
	lives := ms.alives()
	num := rand.Int() % len(lives)
	target := lives[num]
	resp := make([]byte, 1)
	if err := sendReceiveTCP(ctx, target.Addr(), req.encode(), resp); err != nil {
		return false
	}
	if resp[0] != ackRespMsgType {
		return false
	}
	return true
}
