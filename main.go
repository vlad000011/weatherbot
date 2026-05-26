package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"

	"github.com/vlad000011/weatherbot/clients/gemini"
	"github.com/vlad000011/weatherbot/clients/weatherapi"
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

	weatherClient := weatherapi.NewClient(os.Getenv("WEATHERAPI_KEY"))

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
				handleWeatherRequest(bot, callback.Message.Chat.ID, city, weatherClient, geminiClient)

			case strings.HasPrefix(data, "newsuggestion:"):
				city := strings.TrimPrefix(data, "newsuggestion:")
				handleNewSuggestion(bot, callback.Message.Chat.ID, city, weatherClient, geminiClient)
			}
			continue
		}

		if update.Message != nil {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			if update.Message.Text == "/start" {
				startMsg := tgbotapi.NewMessage(update.Message.Chat.ID,
					`👋 <b>Привет! Я — Weather Bot</b>

Пиши название города на английском языке.
Пример: Moscow, Tokyo, London, New York`)
				startMsg.ParseMode = "HTML"
				bot.Send(startMsg)
				continue
			}

			handleWeatherRequest(bot, update.Message.Chat.ID, update.Message.Text, weatherClient, geminiClient)
		}
	}
}

// Основная функция показа погоды
func handleWeatherRequest(bot *tgbotapi.BotAPI, chatID int64, city string, wc *weatherapi.Client, gc *gemini.GeminiClient) {
	city = strings.TrimSpace(city)
	lowCity := strings.ToLower(city)

	// Жёсткий фильтр бреда
	if len(city) < 3 {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Слишком короткое название."))
		return
	}

	badWords := []string{"piska", "dick", "cock", "penis", "fuck", "shit", "bitch", "cunt", "хуй", "пизд", "еб"}
	for _, bad := range badWords {
		if strings.Contains(lowCity, bad) {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Некорректный запрос."))
			return
		}
	}

	weather, err := wc.GetWeather(city)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Город не найден. Пиши правильное английское название.")
		bot.Send(msg)
		return
	}

	// Улучшенная проверка совпадения
	input := strings.ToLower(strings.ReplaceAll(city, "-", " "))
	found := strings.ToLower(strings.ReplaceAll(weather.CityName, "-", " "))

	if !strings.Contains(found, input) && !strings.Contains(input, found) {
		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("❌ Город «%s» не найден.\n\nAPI вернул: %s", city, weather.CityName))
		bot.Send(msg)
		return
	}

	// === Выводим погоду ===
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

	// Gemini рекомендация
	// Gemini рекомендация
	suggestion, err := gc.SuggestClothes(weather, weather.CityName)
	if err != nil {
		log.Printf("=== GEMINI ERROR === City: %s | Error: %v", weather.CityName, err)
		suggestionText := "👕 <b>Что надеть сегодня:</b>\n\nИзвини, сейчас не смог получить рекомендацию 😔"
		suggestionMsg := tgbotapi.NewMessage(chatID, suggestionText)
		suggestionMsg.ParseMode = "HTML"
		bot.Send(suggestionMsg)
	} else {
		suggestionText := "👕 <b>Что надеть сегодня:</b>\n\n" + suggestion
		suggestionMsg := tgbotapi.NewMessage(chatID, suggestionText)
		suggestionMsg.ParseMode = "HTML"
		sentMsg, _ := bot.Send(suggestionMsg)

		// Кнопки
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔄 Обновить", "refresh:"+weather.CityName),
				tgbotapi.NewInlineKeyboardButtonData("👕 Ещё совет", "newsuggestion:"+weather.CityName),
			),
		)
		editMsg := tgbotapi.NewEditMessageReplyMarkup(chatID, sentMsg.MessageID, keyboard)
		bot.Send(editMsg)
	}
}

// Проверка наличия русских букв
func containsRussianLetters(s string) bool {
	for _, r := range s {
		if r >= 'а' && r <= 'я' || r >= 'А' && r <= 'Я' {
			return true
		}
	}
	return false
}

// Проверка похожести городов
func isSimilarCity(input, found string) bool {
	input = strings.ToLower(strings.TrimSpace(input))
	found = strings.ToLower(strings.TrimSpace(found))
	return strings.Contains(found, input) || strings.Contains(input, found)
}

// handleNewSuggestion
func handleNewSuggestion(bot *tgbotapi.BotAPI, chatID int64, city string, wc *weatherapi.Client, gc *gemini.GeminiClient) {
	thinking := tgbotapi.NewMessage(chatID, "🤔 Думаю над новым советом...")
	bot.Send(thinking)

	weather, err := wc.GetWeather(city)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Не удалось получить погоду"))
		return
	}

	suggestion, err := gc.SuggestClothes(weather, weather.CityName)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Gemini не ответил"))
		return
	}

	suggestionText := "👕 <b>Новый совет по одежде:</b>\n\n" + suggestion
	suggestionMsg := tgbotapi.NewMessage(chatID, suggestionText)
	suggestionMsg.ParseMode = "HTML"
	bot.Send(suggestionMsg)
}

func getWeatherEmoji(desc string) string {
	desc = strings.ToLower(desc)
	if strings.Contains(desc, "clear") || strings.Contains(desc, "sunny") {
		return "☀️"
	} else if strings.Contains(desc, "cloud") {
		return "☁️"
	} else if strings.Contains(desc, "rain") {
		return "🌧️"
	} else if strings.Contains(desc, "snow") {
		return "❄️"
	}
	return "🌤️"
}
