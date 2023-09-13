package live

import (
	"fmt"
	"github.com/Akegarasu/blivedm-go/client"
	"github.com/Akegarasu/blivedm-go/message"
	log "github.com/sirupsen/logrus"
)

func WsInit(roomId int) {
	c := client.NewClient(roomId)
	c.SetCookie("")
	c.OnLive(func(live *message.Live) {
		go DoRecord(c.RoomID)
	})
	c.OnDanmaku(func(danmaku *message.Danmaku) {
		if danmaku.Type == message.EmoticonDanmaku {
		} else {
			fmt.Printf("[%d] %s: %s\n", danmaku.Sender.Uid, danmaku.Sender.Uname, danmaku.Content)
		}
	})

	err := c.Start()
	if err != nil {
		log.Fatal(err)
	}
}
