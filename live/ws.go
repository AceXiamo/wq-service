package live

import (
	"fmt"
	"github.com/AceXiamo/blivedm-go/client"
	"github.com/AceXiamo/blivedm-go/message"
	log "github.com/sirupsen/logrus"
)

func WsInit(roomId int) {
	c := client.NewClient(roomId)
	c.SetCookie("innersign=0; buvid3=89F570D1-D749-CF39-6A12-BE76BE61D53D84804infoc; b_nut=1694670984; i-wanna-go-back=-1; b_ut=7; header_theme_version=undefined; home_feed_column=5; browser_resolution=1682-1035; buvid4=4384B3A0-4F6C-66F3-082B-81995E56978E85417-023091413-7EM3j908NRxUR5Sh9HbJAg%3D%3D; SESSDATA=b7b8e0a0%2C1710223016%2C31493%2A91CjBreQtVgqMrN67z4in9okuj0lmbxy8rzM2c1qdTFU2pqqm2nWg1dvFobEvS4ja3evMSVmtOemNLUjBySWkyS0lZcVdKSVMzUlFab2ZzcjludVJ5Tk1feG9hOHR0cGpuMXdHLTQxUFJNREw3dFU1UWtYd3pEbzlmQ2c2dUU0NUcwMnhWX0FmczZnIIEC; bili_jct=9c0c3c1cdcaa9f091e8f6e158d439628; DedeUserID=294797622; DedeUserID__ckMd5=816d40f025ba6c9b; sid=pop4daie; LIVE_BUVID=AUTO2216946710371855")
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
