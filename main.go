package main

import (
	log "github.com/sirupsen/logrus"
	"wq-service/core"
	"wq-service/cos"
	"wq-service/live"
)

func main() {
	core.InitLogger()
	log.SetLevel(log.InfoLevel)
	cos.Init()
	go live.WsInit(25081972)

	select {}
}
