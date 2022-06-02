package main

import (
	"roller-go/config"
	"roller-go/roller"
)

func main() {
	cfg := config.InitConfig("config.toml")
	r := roller.NewRoller(cfg)
	defer r.Close()
	r.Run()
}
