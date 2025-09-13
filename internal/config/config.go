package config

import (
	"flag"
	"log"
	"os"
)

type Config struct {
	ServerAddr  string
	DatabaseDSN string
	AccrualAddr string
}

func MustLoad() *Config {
	var a, d, r string
	defaultAddr := ":8080"

	flag.StringVar(&a, "a", defaultAddr, "address to serve")
	flag.StringVar(&d, "d", "", "database dsn")
	flag.StringVar(&r, "r", "", "accrual system address")
	flag.Parse()

	envAddr := os.Getenv("RUN_ADDRESS")
	if envAddr != "" && a == defaultAddr {
		a = envAddr
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

	return &Config{
		ServerAddr:  a,
		DatabaseDSN: d,
		AccrualAddr: r,
	}
}
