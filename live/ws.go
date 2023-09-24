package live

import (
	"github.com/AceXiamo/blivedm-go/client"
	"github.com/AceXiamo/blivedm-go/message"
	"sync"
	"wq-service/core"
)

func WsInit(roomId int) {
	var recWait *sync.WaitGroup
	c := client.NewClient(roomId)
	c.SetCookie("innersign=0; buvid3=97561807-211A-540C-E3DD-C200FC96C96790512infoc; b_nut=1695294290; i-wanna-go-back=-1; b_ut=7; b_lsid=9B310CC103_18AB769FD74; _uuid=2DC56B5D-C6510-EBF10-FB4C-454843109377991322infoc; header_theme_version=undefined; buvid_fp=9c10ace3750b7383ddd0cb507769b9b8; home_feed_column=5; browser_resolution=1745-845; buvid4=64D58A5E-89A5-F48B-8D9B-0EEB402AEA3491366-023092119-0mOPZbinCsODOwiPFAHr0w%3D%3D; SESSDATA=36c6a3b1%2C1710846326%2Ce7500%2A91CjDqXvB35UDvU1o42juYpfbCoFTf5sHP0LmZJD6vNGLXC7JrgqE1zmCAqF6fvDzSk-sSVk9uZEtNWElhb0RGX0diQjdxSzdYclhLZXJlM3V3MU56S0REUjZ3RFhmaE1pelphUm1ZejMydktlUTlEaW9OSkotSnJSRll5T0Q1RXUydGxmMjVTTDVnIIEC; bili_jct=488244f4f6cdf27c3a17f0341431e00f; DedeUserID=294797622; DedeUserID__ckMd5=816d40f025ba6c9b; sid=7brfmna0; PVID=1; bp_video_offset_294797622=839262392332320777; LIVE_BUVID=AUTO9616952943328378")
	c.OnLive(func(live *message.Live) {
		recWait = &sync.WaitGroup{}
		recWait.Add(1)
		go DoRecord(c.RoomID, recWait)
	})
	c.OnDanmaku(func(danmaku *message.Danmaku) {
		core.Log.Infof("💬 [%d] %s: %s", danmaku.Sender.Uid, danmaku.Sender.Uname, danmaku.Content)
	})
	c.RegisterCustomEventHandler("PREPARING", func(s string) {
		if recWait != nil {
			recWait.Done()
		}
		core.Log.Infof("🟡 [推流停止] %d", c.RoomID)
	})

	err := c.Start()
	if err != nil {
		core.Log.Fatal(err)
	}
}
