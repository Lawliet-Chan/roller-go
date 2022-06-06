package config

import (
	"github.com/BurntSushi/toml"
	"github.com/scroll-tech/go-ethereum/log"
)

type Config struct {
	RollerName   string `toml:"roller_name"`
	Secret       []byte `toml:"secret"`
	ScrollUrl    string `toml:"scroll_url"`
	ProverPath   string `toml:"prover_path"`
	StackPath    string `toml:"stack_path"`
	WsTimeoutSec int    `toml:"ws_timeout_sec"`
}

func InitConfig(path string) *Config {
	var cfg *Config
	_, err := toml.DecodeFile(path, cfg)
	if err != nil {
		log.Crit("init config failed", "error", err)
	}
	return cfg
}
