package bot

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleMessage обрабатывает входящие текстовые сообщения
func (b *Bot) handleMessage(message *tgbotapi.Message) {
	// Обработка команды /start
	if message.Command() == "start" {
		b.handleStartCommand(message.Chat.ID, message.From.ID)
		return
	}
	// Здесь можно добавить логику обработки других команд или текстовых сообщений
}

// handleCallback обрабатывает нажатия на инлайн-кнопки
func (b *Bot) handleCallback(query *tgbotapi.CallbackQuery) {
	switch query.Data {
	case "start":
		b.handleStartCommand(query.From.ID, query.From.ID)
	case "get_meal":
		b.handleGetMeal(query.Message.Chat.ID, query.From.ID)
	}
}

// handleStartCommand обрабатывает команду /start или нажатие кнопки start
func (b *Bot) handleStartCommand(chatID int64, userID int64) {
	uID := fmt.Sprintf("%d", userID)
	var welcomeText string
	if userName, isKnown := b.knownUsers[uID]; isKnown {
		welcomeText = fmt.Sprintf("👋 С возвращением, %s! Нажми кнопку, чтобы получить следующий прием пищи", userName)
	} else {
		welcomeText = "⚠️ Извините, но я вас не знаю. Обратитесь к администратору для получения доступа."
	}

	msg := tgbotapi.NewMessage(chatID, welcomeText)

	if _, isKnown := b.knownUsers[uID]; isKnown {
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Получить следующий прием пищи", "get_meal"),
			),
		)
		msg.ReplyMarkup = keyboard
	}

	b.api.Send(msg)
}

// handleGetMeal обрабатывает получение следующего приема пищи
func (b *Bot) handleGetMeal(chatID int64, userID int64) {
	uID := fmt.Sprintf("%d", userID)
	userName, isKnown := b.knownUsers[uID]
	if !isKnown {
		msg := tgbotapi.NewMessage(chatID, "⚠️ Извините, но я вас не знаю. Обратитесь к администратору для получения доступа.")
		b.api.Send(msg)
		return
	}

	// Приветствие
	welcomeMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("👋 Привет, %s! Сейчас подберу для тебя следующий прием пищи.", userName))
	b.api.Send(welcomeMsg)

	mealResp, err := b.getMealResponse(userName)
	if err != nil {
		b.sendErrorMessage(chatID, err.Error())
		return
	}

	log.Println(mealResp.Meal)
	log.Println(mealResp.ShoppingList)

	message := buildMealMessage(mealResp)
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"

	// Добавляем кнопку для следующего запроса
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Получить следующий прием пищи", "get_meal"),
		),
	)
	msg.ReplyMarkup = keyboard

	b.api.Send(msg)
}
