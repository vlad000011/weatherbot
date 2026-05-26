package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"

	"github.com/vlad000011/weatherbot/clients/gemini"
	"github.com/vlad000011/weatherbot/clients/openmeteo"
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

	meteoClient := openmeteo.NewClient()

	geminiKey := os.Getenv("GEMINI_API_KEY")
	if geminiKey == "" {
		log.Fatal("GEMINI_API_KEY is not set")
	}

	geminiClient, err := gemini.NewClient(geminiKey)
	if err != nil {
		log.Fatalf("Failed to create Gemini client: %v", err)
	}

	for update := range updates {

		if update.CallbackQuery != nil {
			callback := update.CallbackQuery
			data := callback.Data

			answer := tgbotapi.NewCallback(callback.ID, "")
			bot.Request(answer)

			switch {
			case strings.HasPrefix(data, "refresh:"):
				city := strings.TrimPrefix(data, "refresh:")
				handleWeatherRequest(bot, callback.Message.Chat.ID, city, meteoClient, geminiClient)

			case strings.HasPrefix(data, "newsuggestion:"):
				city := strings.TrimPrefix(data, "newsuggestion:")
				handleNewSuggestion(bot, callback.Message.Chat.ID, city, meteoClient, geminiClient)
			}
			continue
		}

		if update.Message != nil {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			if update.Message.Text == "/start" {
				startMsg := tgbotapi.NewMessage(update.Message.Chat.ID, `👋 <b>Привет! Я — Weather Bot</b>

Просто напиши название города.`)
				startMsg.ParseMode = "HTML"
				bot.Send(startMsg)
				continue
			}

			handleWeatherRequest(bot, update.Message.Chat.ID, update.Message.Text, meteoClient, geminiClient)
		}
	}
}

// handleWeatherRequest — основная функция
func handleWeatherRequest(bot *tgbotapi.BotAPI, chatID int64, city string, meteo *openmeteo.Client, gemini *gemini.GeminiClient) {
	lat, lon, cityName, err := meteo.GetCoordinates(city)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Город не найден.")
		bot.Send(msg)
		return
	}

	weather, err := meteo.GetWeather(lat, lon)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Не удалось получить погоду")
		bot.Send(msg)
		return
	}
	weather.CityName = cityName

	weatherText := fmt.Sprintf(`🌍 <b>Погода в %s</b>

%s
Температура: <b>%.1f°C</b>
Ощущается как: <b>%.1f°C</b>
Влажность: <b>%d%%</b>
Ветер: <b>%.1f м/с</b>

%s`,
		weather.CityName,
		getWeatherEmoji(weather.Description),
		weather.Temp,
		weather.FeelsLike,
		weather.Humidity,
		weather.WindSpeed,
		weather.Description,
	)

	msg := tgbotapi.NewMessage(chatID, weatherText)
	msg.ParseMode = "HTML"
	bot.Send(msg)

	suggestion, err := gemini.SuggestClothes(weather, cityName)
	if err != nil {
		log.Printf("Gemini error: %v", err)
	} else {
		suggestionText := "👕 <b>Что надеть сегодня:</b>\n\n" + suggestion
		suggestionMsg := tgbotapi.NewMessage(chatID, suggestionText)
		suggestionMsg.ParseMode = "HTML"
		sentMsg, _ := bot.Send(suggestionMsg)

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔄 Обновить", "refresh:"+cityName),
				tgbotapi.NewInlineKeyboardButtonData("👕 Ещё совет", "newsuggestion:"+cityName),
			),
		)
		editMsg := tgbotapi.NewEditMessageReplyMarkup(chatID, sentMsg.MessageID, keyboard)
		bot.Send(editMsg)
	}
}

func handleNewSuggestion(bot *tgbotapi.BotAPI, chatID int64, city string, meteo *openmeteo.Client, gemini *gemini.GeminiClient) {
	thinking := tgbotapi.NewMessage(chatID, "🤔 Думаю над новым советом...")
	bot.Send(thinking)

	lat, lon, cityName, err := meteo.GetCoordinates(city)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Город не найден"))
		return
	}

	weather, err := meteo.GetWeather(lat, lon)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Не удалось получить погоду"))
		return
	}
	weather.CityName = cityName

	suggestion, err := gemini.SuggestClothes(weather, cityName)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Gemini не ответил"))
		return
	}

	suggestionText := "👕 <b>Новый совет:</b>\n\n" + suggestion
	suggestionMsg := tgbotapi.NewMessage(chatID, suggestionText)
	suggestionMsg.ParseMode = "HTML"
	bot.Send(suggestionMsg)
}

func getWeatherEmoji(desc string) string {
	if strings.Contains(desc, "Ясно") {
		return "☀️"
	} else if strings.Contains(desc, "Облачно") || strings.Contains(desc, "Cloud") {
		return "☁️"
	} else if strings.Contains(desc, "Дождь") || strings.Contains(desc, "Rain") {
		return "🌧️"
	} else if strings.Contains(desc, "Снег") {
		return "❄️"
	}
	return "🌤️"
}
