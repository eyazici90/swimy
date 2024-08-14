package swim

import (
	"net"
	"time"
)

func DefaultConfig() *Config {
	return &Config{
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

type Config struct {
	Port              uint16 // binding lister to
	MaxSuspicionCount int
	GossipInterval    time.Duration // duration of gossiping with members, default: 20ms
	GossipRatio       uint8         // min. percentage of gossiping active members concurrently. default: 20 (%20)
	IOTimeout         time.Duration
	OnJoin, OnLeave   func(addr net.Addr)
}

func setDefaults(ptr **Config) {
	if *ptr == nil {
		*ptr = DefaultConfig()
	}
}
