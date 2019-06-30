package dkvs

import "time"

type config struct {
	retriesCount   int
	retriesDelayMs time.Duration
}

var defaultConfig = &config{
	retriesCount:   3,
	retriesDelayMs: 10000,
}

var encoding = "application/json"
