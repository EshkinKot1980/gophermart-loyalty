package config

import (
	"errors"
	"flag"
	"log"
	"os"
	"strconv"

	accrual "github.com/EshkinKot1980/gophermart-loyalty/internal/accrual/config"
)

const (
	defaultAddr         = ":8080"
	defaultSecret       = "J3gdkl8v3hPJ8" // нужен для того, чтобы приемочные тесты прошли
	defaultRateLimit    = 10
	defaultPollInterval = 1
	defaultProcessDelay = 10
	defaultRetryCount   = 3
)

var ErrNotNaturalNumber = errors.New("the value must be a natural number")

type Config struct {
	ServerAddr  string
	DatabaseDSN string
	JWTsecret   string
	AccrualGfg  *accrual.Config
}

func MustLoad() *Config {
	var a, d, r, s string

	rateLimit := newNatural(defaultRateLimit)
	pollInterval := newNatural(defaultPollInterval)
	processDelay := newNatural(defaultProcessDelay)
	retryCount := newNatural(defaultRetryCount)

	flag.StringVar(&a, "a", defaultAddr, "address to serve")
	flag.StringVar(&d, "d", "", "database dsn")
	flag.StringVar(&r, "r", "", "accrual system address")
	flag.StringVar(&s, "s", defaultSecret, "jwt secret")
	flag.Var(rateLimit, "rl", "accrual system rate limit, limit of simultaneous requests")
	flag.Var(pollInterval, "pi", "accrual system db poll interval in seconds")
	flag.Var(processDelay, "pd", "accrual system process delay in seconds")
	flag.Var(retryCount, "rc", "accrual system retry count for unregistered orders")
	flag.Parse()

	envAddr := os.Getenv("RUN_ADDRESS")
	if envAddr != "" && a == defaultAddr {
		a = envAddr
	}

	envSecret := os.Getenv("JWT_SECRET")
	if envSecret != "" && s == defaultSecret {
		s = envSecret
	}

	envDSN := os.Getenv("DATABASE_URI")
	if envDSN != "" && d == "" {
		d = envDSN
	}
	if d == "" {
		log.Fatal("database dsn is required")
	}

	envAccrualAddr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS")
	if envAccrualAddr != "" && r == "" {
		r = envAccrualAddr
	}
	if r == "" {
		log.Fatal("accural system address is required")
	}

	envPollInterval := os.Getenv("ACCRUAL_DB_POLL_INTERVAL")
	if envPollInterval != "" && !pollInterval.isSet {
		err := pollInterval.Set(envPollInterval)
		if err != nil {
			log.Fatal(err)
		}
	}

	envProcessDelay := os.Getenv("ACCRUAL_PROCESS_DELAY")
	if envProcessDelay != "" && !processDelay.isSet {
		err := processDelay.Set(envProcessDelay)
		if err != nil {
			log.Fatal(err)
		}
	}

	envRetryCount := os.Getenv("ACCRUAL_NOT_REGISTER_RETRY_COUNT")
	if envRetryCount != "" && !retryCount.isSet {
		err := retryCount.Set(envRetryCount)
		if err != nil {
			log.Fatal(err)
		}
	}

	envRateLimit := os.Getenv("ACCRUAL_RATE_LIMIT")
	if envRateLimit != "" && !rateLimit.isSet {
		err := rateLimit.Set(envRateLimit)
		if err != nil {
			log.Fatal(err)
		}
	}

	return &Config{
		ServerAddr:  a,
		DatabaseDSN: d,
		JWTsecret:   s,
		AccrualGfg: &accrual.Config{
			AccrualAddr:         r,
			RateLimit:           rateLimit.value,
			PollInterval:        pollInterval.value,
			ProcessDelay:        processDelay.value,
			UnregisteredRetries: retryCount.value,
		},
	}
}

type natural struct {
	value uint64
	isSet bool
}

func newNatural(v uint64) *natural {
	if v == 0 {
		log.Fatal("attempt to initi a natural number by zero")
	}

	return &natural{value: v}
}

func (n *natural) String() string {
	return strconv.FormatUint(n.value, 10)
}

func (n *natural) Set(flagValue string) error {
	v, err := strconv.ParseUint(flagValue, 10, 64)
	if err != nil {
		return ErrNotNaturalNumber
	}
	if v == 0 {
		return ErrNotNaturalNumber
	}
	n.value = v
	n.isSet = true
	return nil
}
