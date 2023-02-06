package faucet

import (
	"math/big"
	"time"
)

type Config struct {
	PrivateKey           string
	Provider             string
	Delay                time.Duration
	CacheCleanupDuration time.Duration
	RPCTimeout           time.Duration
	AddressQueryKey      string
	FallbackGasPrice     *big.Int
	Amount               *big.Int
	AmountPerDay         *big.Int
	ChainID              *big.Int
}

func useDefaults(cfg *Config) *Config {
	if cfg == nil {
		cfg = &Config{}
	}

	if cfg.Delay == 0 {
		cfg.Delay = time.Hour * 12
	}

	if cfg.CacheCleanupDuration == 0 {
		cfg.CacheCleanupDuration = time.Minute * 2
	}

	if cfg.RPCTimeout == 0 {
		cfg.RPCTimeout = time.Second * 10
	}

	if cfg.AddressQueryKey == "" {
		cfg.AddressQueryKey = "address"
	}

	cfg.FallbackGasPrice = useNilValue(cfg.ChainID, big.NewInt(1e9))
	cfg.Amount = useNilValue(cfg.Amount, big.NewInt(1e16))
	cfg.AmountPerDay = useNilValue(cfg.AmountPerDay, big.NewInt(5))

	return cfg
}

func useNilValue(t *big.Int, fallback *big.Int) *big.Int {
	if t != nil {
		return t
	}
	return fallback
}
