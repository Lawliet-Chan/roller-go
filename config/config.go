package config

import (
	"github.com/BurntSushi/toml"
	"github.com/scroll-tech/go-ethereum/log"
)

type Config struct {
	RollerName       string `toml:"roller_name"`
	SecretKey        string `toml:"secret_key"`
	ScrollUrl        string `toml:"scroll_url"`
	ProverSocketPath string `toml:"prover_socket_path"`
	StackPath        string `toml:"stack_path"`
	WsTimeoutSec     int    `toml:"ws_timeout_sec"`
}

func InitConfig(path string) *Config {
	var cfg *Config
	_, err := toml.DecodeFile(path, cfg)
	if err != nil {
		log.Crit("init config failed", "error", err)
	}
	return cfg
}
