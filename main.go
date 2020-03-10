package main

import (
	"meeting_bot/internal/config"
	"meeting_bot/internal/log"
)

func main() {
	log := log.InitLogger("meetingbot.log", "info")
	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatal(err)
	}
	
	log.Info("Running service  ", cfg.Network.Host, ":", cfg.Network.Port)
}