package main

import (
	"github.com/Lawliet-Chan/roller-go/config"
	"github.com/Lawliet-Chan/roller-go/roller"
)

func main() {
	cfg := config.InitConfig("config.toml")
	r := roller.NewRoller(cfg)
	defer r.Close()
	r.Run()
}
