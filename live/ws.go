package live

import (
	"github.com/Akegarasu/blivedm-go/client"
	"github.com/Akegarasu/blivedm-go/message"
	log "github.com/sirupsen/logrus"
)

func WsInit(roomId int) {
	c := client.NewClient(roomId)
	c.SetCookie("")
	c.OnLive(func(live *message.Live) {
		log.Infof("🟢 [直播开始] %d", live.Roomid)
		go DoRecord(c.RoomID)
	})

	err := c.Start()
	if err != nil {
		log.Fatal(err)
	}
}
