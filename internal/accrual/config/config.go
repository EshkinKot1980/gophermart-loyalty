package config

type Config struct {
	AccrualAddr         string
	RateLimit           uint64
	ProcessDelay        uint64
	PollInterval        uint64
	UnregisteredRetries uint64
}
