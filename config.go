package swim

import "time"

type Config struct {
	MaxSuspicionCount int
	GossipInterval    time.Duration
	IOTimeout         time.Duration

	OnJoin, OnLeave func(m *Member)
}

func setDefaults(ptr **Config) {
	if *ptr == nil {
		*ptr = &Config{
			MaxSuspicionCount: 5,
			GossipInterval:    time.Millisecond * 20,
			IOTimeout:         time.Millisecond * 100,
			OnJoin: func(m *Member) {
			},
			OnLeave: func(m *Member) {
			},
		}
	}
}
