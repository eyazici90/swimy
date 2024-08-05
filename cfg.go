package swim

import "time"

type Config struct {
	MaxSuspicionCount int
	GossipInterval    time.Duration

	ProbeInterval time.Duration
	ProbeTimeout  time.Duration
}
