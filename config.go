package swim

import "time"

type Config struct {
	MaxSuspicionCount int
	GossipInterval    time.Duration
	IOTimeout         time.Duration
}

func setDefaults(ptr **Config) {
	if *ptr == nil {
		*ptr = &Config{
			MaxSuspicionCount: 5,
			GossipInterval:    time.Millisecond * 20,
			IOTimeout:         time.Millisecond * 100,
		}
	}
}
