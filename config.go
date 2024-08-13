package swim

import (
	"net"
	"time"
)

type Config struct {
	Port              uint16 // binding lister to
	MaxSuspicionCount int
	GossipInterval    time.Duration // duration of gossiping with members
	GossipRatio       uint8         // min. percentage of gossiping active members concurrently. ex: 20 => %20
	IOTimeout         time.Duration
	OnJoin, OnLeave   func(m net.Addr)
}

func setDefaults(ptr **Config) {
	if *ptr == nil {
		*ptr = &Config{
			MaxSuspicionCount: 5,
			GossipRatio:       20,
			GossipInterval:    time.Millisecond * 20,
			IOTimeout:         time.Millisecond * 100,
			OnJoin: func(_ net.Addr) {
			},
			OnLeave: func(_ net.Addr) {
			},
		}
	}
}
