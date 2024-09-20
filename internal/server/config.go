package server

import (
	"math"
	"math/big"
)

type Config struct {
	network      string
	httpPort     int
	interval     int
	payout       *big.Int
	payoutAmount *big.Int
	proxyCount   int
	queueCap     int
}

func NewConfig(network string, httpPort, interval, payout, precision, proxyCount, queueCap int) *Config {
	return &Config{
		network:      network,
		httpPort:     httpPort,
		interval:     interval,
		payout:       big.NewInt(int64(payout)),
		payoutAmount: big.NewInt(int64(payout * int(math.Pow10(precision)))),
		proxyCount:   proxyCount,
		queueCap:     queueCap,
	}
}
