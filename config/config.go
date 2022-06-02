package config

import (
	"github.com/BurntSushi/toml"
	"github.com/scroll-tech/go-ethereum/log"
)

type Config struct {
	RollerName   string
	Secret       []byte
	ScrollUrl    string
	ProverPath   string
	StackPath    string
	WsTimeoutSec int
}

func InitConfig(path string) *Config {
	var cfg *Config
	_, err := toml.DecodeFile(path, cfg)
	if err != nil {
		log.Crit("init config failed", "error", err)
	}
	return cfg
}
