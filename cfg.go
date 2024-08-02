package swim

import "time"

type Config struct {
	MaxSuspicionCount int

	ProbeInterval time.Duration
	ProbeTimeout  time.Duration
}
