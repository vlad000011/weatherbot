package main

import (
	"fmt"
	"log"
	"math"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/vlad000011/weatherbot/clients/openweather"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	owClient := openweather.New(os.Getenv("OPENWEATHERAPI_KEY"))

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			coordinates, err := owClient.Coordinates(update.Message.Text)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "не смогли получить координаты")
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)
				continue
			}
			weather, err := owClient.Weather(coordinates.Lat, coordinates.Lon)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "не смогли получить погоду в вашей местности")
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)
				continue
			}

			msg := tgbotapi.NewMessage(
				update.Message.Chat.ID,
				fmt.Sprintf("температура в %s:%g", update.Message.Text, math.Round(weather.Temp)))
			msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)
		}
	}
}
