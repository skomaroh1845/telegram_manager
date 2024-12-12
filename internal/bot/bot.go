package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Bot представляет собой телеграм-бота со всеми необходимыми зависимостями.
type Bot struct {
	api            *tgbotapi.BotAPI
	menuServiceURL string
	knownUsers     map[string]string
}

// New создает новый экземпляр бота.
// Принимает токен доступа к боту, URL сервиса меню и словарь известных пользователей.
func New(token string, menuServiceURL string, knownUsers map[string]string) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать бота: %w", err)
	}

	return &Bot{
		api:            api,
		menuServiceURL: menuServiceURL,
		knownUsers:     knownUsers,
	}, nil
}

// Start запускает бота и обрабатывает входящие обновления.
// При запуске отправляет сообщение известным пользователям, при остановке - уведомляет об этом.
func (b *Bot) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Приветственное сообщение при запуске
	b.notifyAllUsersOnStart()

	updates := b.api.GetUpdatesChan(u)
	for update := range updates {
		// Обработка входящих обновлений (сообщений, команд, колбэков)
		if update.Message != nil {
			b.handleMessage(update.Message)
		} else if update.CallbackQuery != nil {
			b.handleCallback(update.CallbackQuery)
		}
	}

	// Уведомление об остановке бота
	b.notifyAllUsersOnStop()

	return nil
}

// notifyAllUsersOnStart отправляет всем известным пользователям сообщение о старте бота.
func (b *Bot) notifyAllUsersOnStart() {
	for chatID, userName := range b.knownUsers {
		id, err := strconv.ParseInt(chatID, 10, 64)
		if err != nil {
			log.Println("Ошибка преобразования chatID:", err)
			continue
		}
		message := fmt.Sprintf("🍖 %s, Бот запущен.\nДа начнется массонабор!", userName)
		msg := tgbotapi.NewMessage(id, message)

		// Кнопка для старта
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Старт!", "start"),
			),
		)
		msg.ReplyMarkup = keyboard

		b.api.Send(msg)
	}
}

// notifyAllUsersOnStop отправляет всем известным пользователям сообщение о стопе бота.
func (b *Bot) notifyAllUsersOnStop() {
	for chatID, userName := range b.knownUsers {
		id, err := strconv.ParseInt(chatID, 10, 64)
		if err != nil {
			log.Println("Ошибка преобразования chatID:", err)
			continue
		}
		message := fmt.Sprintf("🍖 %s, Бот остановлен.\nДа начнется сушка!", userName)
		msg := tgbotapi.NewMessage(id, message)
		b.api.Send(msg)
	}
}

// sendErrorMessage отправляет сообщение об ошибке пользователю
// и кнопкой "Попробовать снова" для повторной попытки.
func (b *Bot) sendErrorMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("❌ Ошибка: %s", text))
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Попробовать снова", "get_meal"),
		),
	)
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

// getMealResponse делает запрос к сервису меню и возвращает MealResponse.
func (b *Bot) getMealResponse(userName string) (*MealResponse, error) {
	query := fmt.Sprintf("%s/api/v1/menus/getMeal?user_id=%s", b.menuServiceURL, userName)
	log.Println(time.Now().String() + " -> " + query)
	resp, err := http.Get(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса к сервису меню: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("статус ответа не ОК\n%s", string(body))
	}

	var mealResp MealResponse
	if err := json.NewDecoder(resp.Body).Decode(&mealResp); err != nil {
		return nil, fmt.Errorf("ошибка при разборе данных: %w", err)
	}
	return &mealResp, nil
}
