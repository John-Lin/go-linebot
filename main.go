package main

import (
	"log"
	"strings"
	"net/http"
	"os"
	"fmt"
	"time"
	"encoding/json"
	"io/ioutil"
	"strconv"

	"github.com/line/line-bot-sdk-go/linebot"
	"gopkg.in/gin-gonic/gin.v1"
)

type Response struct {
  // The right side is the name of the JSON variable
	Success       bool                 `json:"success"`
	Timestamp     int64                `json:"timestamp"`
	Source        string               `json:"source"`
	Quotes        map[string]float64   `json:"quotes"`
}

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
				fmt.Println(err)
				c.AbortWithError(400, err)
			} else {
				fmt.Println(err)
				c.AbortWithError(500, err)
			}
			return
		}

		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					log.Printf("User ID is %v\n", event.Source.UserID)

					if strings.HasPrefix(message.Text, "/") {
						if IsLetter(message.Text[1:4]) {
							r := convertCurrency()
							if r.Success == true {
								code := strings.ToUpper(message.Text[1:4])

								if r.Quotes["USD" + code] == 0 {
									_, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("沒有這個外匯代號")).Do()
									if err != nil { log.Print(err) }
								} else {
									// result := time.Unix(r.Timestamp, 0).String() + "\n" + FloatToString(r.Quotes["USD" + code])
									result := "USD/" + code + "  " + FloatToString(r.Quotes["USD" + code])
									_, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(result)).Do()
									if err != nil { log.Print(err) }
								}
							} else {
								_, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("Service Unreachable!")).Do()
								if err != nil { log.Print(err) }
								log.Printf("Service Unreachable!")
							}

						}
					}
					// _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.Text)).Do()
					// if err != nil { log.Print(err) }

				}
			}
		}
	})

	router.Run(":" + port)
}

func convertCurrency() *Response {
	var (
		queryString  string
		apiKey       string
		err          error
		resp         Response
		response     *http.Response
		body         []byte
	)

	apiKey = os.Getenv("currencylayerAPIKey")

	// Euro by default. Setting the base parameter in your request.
	queryString = "?access_key=" + apiKey + "&source=USD&format=1"

	// Use api.fixer.io to get a JSON response
	response, err = http.Get("http://apilayer.net/api/live" + queryString)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()

	// response.Body() is a reader type. We have
	// to use ioutil.ReadAll() to read the data
	// in to a byte slice(string)
	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
	}

	// Unmarshal the JSON byte slice to a GeoIP struct
	err = json.Unmarshal(body, &resp)
	if err != nil {
		fmt.Println(err)
	}

	// Everything accessible in struct now
	return &resp
}

func IsLetter(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
			return false
		}
	}
	return true
}

func FloatToString(f float64) string {
    // to convert a float number to a string
    return strconv.FormatFloat(f, 'f', 3, 64)
}
