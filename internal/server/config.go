package server

import (
	"math/big"
)

type Config struct {
	network    string
	httpPort   int
	interval   int
	payout     *big.Int
	payoutNano *big.Int
	proxyCount int
	queueCap   int
}

func NewConfig(network string, httpPort, interval, payout, proxyCount, queueCap int) *Config {
	return &Config{
		network:    network,
		httpPort:   httpPort,
		interval:   interval,
		payout:     big.NewInt(int64(payout)),
		payoutNano:     big.NewInt(int64(payout * 1e9)),
		proxyCount: proxyCount,
		queueCap:   queueCap,
	}
}
