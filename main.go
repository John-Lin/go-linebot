package main

import (
	"log"
	"net/http"
	"os"
	"fmt"
	"time"

	"github.com/line/line-bot-sdk-go/linebot"
	"gopkg.in/gin-gonic/gin.v1"
)

func main() {
	port := os.Getenv("PORT")
	// port := "9000"

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.New()
	router.Use(gin.Logger())


	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "PONG")
	})

	router.POST("/callback", func(c *gin.Context) {
		client := &http.Client{Timeout: time.Duration(15 * time.Second)}

		bot, err := linebot.New(
			os.Getenv("CHANNEL_SECRET"),
			os.Getenv("CHANNEL_TOKEN"),
			linebot.WithHTTPClient(client))
		if err != nil {
				fmt.Println(err)
				return
		}

		events, err := bot.ParseRequest(c.Request)

		if err != nil {
			if err == linebot.ErrInvalidSignature {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(500)
			}
			return
		}

		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					_, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.ID+":"+message.Text+" OK!")).Do()
					if err != nil { log.Print(err) }

				}
			}
		}
	})

	router.Run(":" + port)
}
