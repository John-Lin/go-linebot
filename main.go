package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/line/line-bot-sdk-go/linebot"
	"gopkg.in/gin-gonic/gin.v1"
)

func main() {
	app, err := NewCurrencyBot(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)
	if err != nil {
		log.Fatal(err)
	}

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
		events, err := app.bot.ParseRequest(c.Request)

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
			switch event.Type {
			case linebot.EventTypeMessage:
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					log.Printf("User ID is %v\n", event.Source.UserID)
					match, _ := regexp.MatchString("([a-zA-Z]{3})/([a-zA-Z]{3})", message.Text)

					if match == true {
						r, _ := regexp.Compile("([a-zA-Z]{3})/([a-zA-Z]{3})")
						res := r.FindAllStringSubmatch(message.Text, -1)
						sourceCurrencySymbol := strings.ToUpper(res[0][1])
						targetCurrencySymbol := strings.ToUpper(res[0][2])
						convertResult := app.convertCurrency()

						if convertResult.Success == true {
							sourceCurrencyQuote := convertResult.Quotes["USD"+sourceCurrencySymbol]
							targetCurrencyQuote := convertResult.Quotes["USD"+targetCurrencySymbol]

							if checkValidCurrency(sourceCurrencyQuote) && checkValidCurrency(targetCurrencyQuote) == true {
								if err := app.replyText(event.ReplyToken, "查無此匯率代號"); err != nil {
									log.Print(err)
								}
							}

							calculatedQuote := targetCurrencyQuote / sourceCurrencyQuote
							result := sourceCurrencySymbol + "/" + targetCurrencySymbol + "  " + FloatToString(calculatedQuote)

							if err := app.replyText(event.ReplyToken, result); err != nil {
								log.Print(err)
							}
						}
					} else {
						// Not match! Might input a invalid currency symbol
						if err := app.replyText(event.ReplyToken, "匯率代號輸入錯誤"); err != nil {
							log.Print(err)
						}
					}

				default:
					log.Printf("Unknown message: %v", message)
				}
			default:
				log.Printf("Unknown event: %v", event)
			}

		}
	})

	router.Run(":" + port)
}

// CurrencyBot app
type CurrencyBot struct {
	bot *linebot.Client
}

type Response struct {
	// The right side is the name of the JSON variable
	Success   bool               `json:"success"`
	Timestamp int64              `json:"timestamp"`
	Source    string             `json:"source"`
	Quotes    map[string]float64 `json:"quotes"`
}

// NewCurrencyBot function
func NewCurrencyBot(channelSecret, channelToken string) (*CurrencyBot, error) {
	client := &http.Client{Timeout: time.Duration(15 * time.Second)}
	bot, err := linebot.New(
		channelSecret,
		channelToken,
		linebot.WithHTTPClient(client),
	)
	if err != nil {
		return nil, err
	}

	return &CurrencyBot{bot: bot}, nil
}

func (app *CurrencyBot) convertCurrency() *Response {
	var (
		queryString string
		apiKey      string
		err         error
		resp        Response
		response    *http.Response
		body        []byte
	)

	apiKey = os.Getenv("currencylayerAPIKey")

	// Setting the base parameter in your request.
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

func (app *CurrencyBot) replyText(replyToken, text string) error {
	if _, err := app.bot.ReplyMessage(
		replyToken,
		linebot.NewTextMessage(text),
	).Do(); err != nil {
		return err
	}
	return nil
}

func FloatToString(f float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(f, 'f', 5, 64)
}

func checkValidCurrency(f float64) bool {
	if f != 0 {
		return true
	}
	return false
}
