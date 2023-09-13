package main

import (
	_ "github.com/Akegarasu/blivedm-go/utils"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
	"wq-service/cos"
	"wq-service/live"
)

func main() {
	log.SetLevel(log.InfoLevel)
	cos.Init()
	go live.WsInit(2696287)
	go countTask()

	r := gin.Default()
	router(r)
	err := r.Run()
	if err != nil {
		return
	}
}

// router
// @Description: router
// @param r
func router(r *gin.Engine) {
	r.GET("/lives", func(c *gin.Context) {
		next := c.Query("next")
		c.JSON(200, live.ListRecord(next))
	})
}

// countTask
// @Description: countTask
func countTask() {
	log.Infof("⚙️ [Task] Start count task")
	live.LoadCount()
	c := cron.New()
	err := c.AddFunc("0 0 6 * * ?", func() {
		live.LoadCount()
	})
	if err != nil {
		return
	}
	c.Start()
}
