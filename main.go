package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
)

func main() {
	r := gin.Default()
	// routes here
	r.POST("/update/notification", func(c *gin.Context) {

		var notificationUpdate NotificationUpdate
		err := c.ShouldBindJSON(&notificationUpdate)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"ErrorMsg":  "欄位錯誤",
				"ErrorFlag": "3",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"ErrorMsg":  notificationUpdate.UserToken,
			"ErrorFlag": "0",
		})

	})
	r.Run()
}

type UpdateNotification struct {
	Name   string `json:"name"`
	Switch bool   `json:"switch"`
}

func SetFCMSetting() {
	timezone, _ := time.LoadLocation("Asia/Taipei")
	s := gocron.NewScheduler(timezone)

	// 每3秒执行一次
	// s.Every(3).Seconds().Do(TriggerFCM)
	s.Every(1).Day().At("08:00").Do(TriggerFCM)
	s.StartBlocking()
}

func TriggerFCM() {
	fmt.Println("八點執行FCM")
}

type NotificationUpdate struct {
	UserToken         string `json:"UserToken"`
	NotificationValue bool   `json:"NotificationValue"`
}
