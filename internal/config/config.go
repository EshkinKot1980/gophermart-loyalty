package config

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	accrual "github.com/EshkinKot1980/gophermart-loyalty/internal/accrual/config"
)

var ErrNotNaturalNumber = errors.New("value must be a natural number")

type Config struct {
	ServerAddr  string
	DatabaseDSN string
	JWTsecret   string
	AccrualGfg  *accrual.Config
}

func Load() (*Config, error) {
	var (
		serverAddr   = newStringVal(":8080")
		dbDSN        = newStringVal("")
		accrualAddr  = newStringVal("")
		secret       = newStringVal("J3gdkl8v3hPJ8")
		rateLimit    = newNaturalVal(10)
		pollInterval = newNaturalVal(1)
		processDelay = newNaturalVal(10)
		retryCount   = newNaturalVal(3)
	)

	flagSet := flag.NewFlagSet("", flag.ContinueOnError)

	flagSet.Var(serverAddr, "a", "address to serve")
	flagSet.Var(dbDSN, "d", "database dsn")
	flagSet.Var(accrualAddr, "r", "accrual system address")
	flagSet.Var(secret, "s", "jwt secret")
	flagSet.Var(rateLimit, "rl", "accrual system rate limit, limit of simultaneous requests")
	flagSet.Var(pollInterval, "pi", "accrual system db poll interval in seconds")
	flagSet.Var(processDelay, "pd", "accrual system process delay in seconds")
	flagSet.Var(retryCount, "rc", "accrual system retry count for unregistered orders")

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		return &Config{}, fmt.Errorf("failed to parse flags")
	}

	envAddr, ok := os.LookupEnv("RUN_ADDRESS")
	if ok && !serverAddr.isset {
		serverAddr.Set(envAddr)
	}

	envSecret, ok := os.LookupEnv("JWT_SECRET")
	if ok && !secret.isset {
		secret.Set(envSecret)
	}

	envDSN, ok := os.LookupEnv("DATABASE_URI")
	if ok && !dbDSN.isset {
		dbDSN.Set(envDSN)
	}
	if dbDSN.value == "" {
		return &Config{}, fmt.Errorf("database dsn is required")
	}

	envAccrualAddr, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS")
	if ok && !accrualAddr.isset {
		accrualAddr.Set(envAccrualAddr)
	}
	if accrualAddr.value == "" {
		return &Config{}, fmt.Errorf("accural system address is required")
	}

	envPollInterval, ok := os.LookupEnv("ACCRUAL_DB_POLL_INTERVAL")
	if ok && !pollInterval.isSet {
		err := pollInterval.Set(envPollInterval)
		if err != nil {
			return &Config{}, fmt.Errorf("ACCRUAL_DB_POLL_INTERVAL %w", err)
		}
	}

	envProcessDelay, ok := os.LookupEnv("ACCRUAL_PROCESS_DELAY")
	if ok && !processDelay.isSet {
		err := processDelay.Set(envProcessDelay)
		if err != nil {
			return &Config{}, fmt.Errorf("ACCRUAL_PROCESS_DELAY %w", err)
		}
	}

	envRetryCount, ok := os.LookupEnv("ACCRUAL_NOT_REGISTER_RETRY_COUNT")
	if ok && !retryCount.isSet {
		err := retryCount.Set(envRetryCount)
		if err != nil {
			return &Config{}, fmt.Errorf("ACCRUAL_NOT_REGISTER_RETRY_COUNT %w", err)
		}
	}

	envRateLimit, ok := os.LookupEnv("ACCRUAL_RATE_LIMIT")
	if ok && !rateLimit.isSet {
		err := rateLimit.Set(envRateLimit)
		if err != nil {
			return &Config{}, fmt.Errorf("ACCRUAL_RATE_LIMIT %w", err)
		}
	}

	config := Config{
		ServerAddr:  serverAddr.value,
		DatabaseDSN: dbDSN.value,
		JWTsecret:   secret.value,
		AccrualGfg: &accrual.Config{
			AccrualAddr:         accrualAddr.value,
			RateLimit:           rateLimit.value,
			PollInterval:        pollInterval.value,
			ProcessDelay:        processDelay.value,
			UnregisteredRetries: retryCount.value,
		},
	}

	return &config, nil
}

type stringVal struct {
	value string
	isset bool
}

func newStringVal(v string) *stringVal {
	return &stringVal{value: v}
}

func (s *stringVal) String() string {
	return s.value
}

func (s *stringVal) Set(flagValue string) error {
	s.value = flagValue
	s.isset = true
	return nil
}

type naturalVal struct {
	value uint64
	isSet bool
}

func newNaturalVal(v uint64) *naturalVal {
	if v == 0 {
		log.Fatal("attempt to initi a natural number by zero")
	}

	return &naturalVal{value: v}
}

func (n *naturalVal) String() string {
	return strconv.FormatUint(n.value, 10)
}

func (n *naturalVal) Set(flagValue string) error {
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
