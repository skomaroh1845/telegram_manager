package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MealResponse структура для ответа от сервиса меню
type MealResponse struct {
	Meal struct {
		ID        string   `json:"id"`
		DishIDs   []string `json:"ID_dish"`
		DishName  []string `json:"dishname"`
		Type      string   `json:"type"`
		Recipe    []string `json:"recipe"`
		Nutrition struct {
			Proteins      int `json:"proteins"`
			Fats          int `json:"fats"`
			Carbohydrates int `json:"carbohydrates"`
			Calories      int `json:"calories"`
		} `json:"total_nutrition"`
	} `json:"meal"`
	ShoppingList string `json:"shopping_list"`
}

type Ingredient struct {
	Unit      string  `json:"unit"`
	Amount    float64 `json:"amount"`
	ProductID string  `json:"product_id"`
}

type RecipeData struct {
	Steps       []string     `json:"steps"`
	Ingredients []Ingredient `json:"ingredients"`
}

// Bot структура для работы с ботом
type Bot struct {
	api            *tgbotapi.BotAPI
	menuServiceURL string
	knownUsers     map[string]string
}

// ShoppingListItem represents a single item in the shopping list
type ShoppingListItem struct {
	ID                       string    `json:"id"`
	Name                     string    `json:"name"`
	WeightPerPkg             float64   `json:"weight_per_pkg"`
	Amount                   int       `json:"amount"`
	PricePerPkg              float64   `json:"price_per_pkg"`
	ExpirationDate           time.Time `json:"expiration_date"`
	PresentInFridge          bool      `json:"present_in_fridge"`
	NutritionalValueRelative struct {
		Proteins      int `json:"proteins"`
		Fats          int `json:"fats"`
		Carbohydrates int `json:"carbohydrates"`
		Calories      int `json:"calories"`
	} `json:"nutritional_value_relative"`
}

// ShoppingListResponse represents the shopping list JSON structure
type ShoppingListResponse struct {
	Products []ShoppingListItem `json:"products"`
}

func formatShoppingList(jsonStr string) string {
	var response ShoppingListResponse
	if err := json.Unmarshal([]byte(jsonStr), &response); err != nil {
		log.Println(fmt.Errorf("failed to create bot: %w", err))
		return "❌ Ошибка в формате списка покупок"
	}

	var result strings.Builder
	result.WriteString("🛒 *Список покупок:*\n")

	for _, item := range response.Products {
		// Skip items with empty names, use ID if name is empty
		itemName := item.Name
		if itemName == "" {
			itemName = item.ID
		}
		log.Println(item)

		result.WriteString(fmt.Sprintf("• %s", itemName))
		if item.Amount > 0 {
			result.WriteString(fmt.Sprintf(" (%d г)", item.Amount))
		}
		if item.WeightPerPkg > 0 {
			result.WriteString(fmt.Sprintf(" %.2f кг", item.WeightPerPkg))
		}
		if item.PricePerPkg > 0 {
			result.WriteString(fmt.Sprintf(" - %.2f₽", item.PricePerPkg))
		}
		result.WriteString("\n")
	}

	return result.String()
}

// New создает новый экземпляр бота
func New(token string, menuServiceURL string, knownUsers map[string]string) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	return &Bot{
		api:            api,
		menuServiceURL: menuServiceURL,
		knownUsers:     knownUsers,
	}, nil
}

// Start запускает бот
func (b *Bot) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Отбивка старта бота
	for chatID, userName := range b.knownUsers {
		id, err := strconv.ParseInt(chatID, 10, 64)
		if err != nil {
			fmt.Println("Ошибка преобразования chatID:", err)
			continue
		}
		// приветственное сообщение
		message := fmt.Sprintf("🍖 %s, Бот запущен.\nДа начнется массонабор!", userName)
		msg := tgbotapi.NewMessage(id, message)

		// добвляем кнопку старт
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Старт!", "start"),
			),
		)
		msg.ReplyMarkup = keyboard

		b.api.Send(msg)
	}

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {

		if update.Message == nil && update.CallbackQuery == nil {
			continue
		}

		// Обработка команды /start и кнопки start
		if (update.Message != nil && update.Message.Command() == "start") || (update.CallbackQuery != nil && update.CallbackQuery.Data == "start") {
			var userID string
			if update.Message != nil {
				userID = fmt.Sprintf("%d", update.Message.From.ID)
			} else {
				userID = fmt.Sprintf("%d", update.CallbackQuery.From.ID)
			}

			// для известных пользователей
			var welcomeText string
			if userName, isKnown := b.knownUsers[userID]; isKnown {
				welcomeText = fmt.Sprintf("👋 С возвращением, %s! Нажми кнопку, чтобы получить следующий прием пищи", userName)
			} else {
				welcomeText = "⚠️ Извините, но я вас не знаю. Обратитесь к администратору для получения доступа."
			}

			var msg tgbotapi.MessageConfig
			if update.Message != nil {
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, welcomeText)
			} else {
				msg = tgbotapi.NewMessage(update.CallbackQuery.From.ID, welcomeText)
			}

			// Добавляем кнопку только для известных пользователей
			if _, isKnown := b.knownUsers[userID]; isKnown {
				keyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Получить следующий прием пищи", "get_meal"),
					),
				)
				msg.ReplyMarkup = keyboard
			}

			b.api.Send(msg)
			continue
		}

		// Обработка нажатия кнопки
		if update.CallbackQuery != nil && update.CallbackQuery.Data == "get_meal" {
			// stubJson := `{"products":[{"id":"Apple","name":"Яблоки","weight_per_pkg":1,"amount":5,"price_per_pkg":150,"expiration_date":"0001-01-01T00:00:00Z","present_in_fridge":false,"nutritional_value_relative":{"proteins":0,"fats":0,"carbohydrates":0,"calories":0}}]}`
			// stubMsg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, formatShoppingList(stubJson))

			// Получаем ID пользователя
			userID := fmt.Sprintf("%d", update.CallbackQuery.From.ID)

			// Проверяем, известен ли пользователь
			userName, isKnown := b.knownUsers[userID]
			if !isKnown {
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "⚠️ Извините, но я вас не знаю. Обратитесь к администратору для получения доступа.")
				b.api.Send(msg)
				continue
			}

			// Приветствуем известного пользователя
			welcomeMsg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, fmt.Sprintf("👋 Привет, %s! Сейчас подберу для тебя следующий прием пищи.", userName))
			b.api.Send(welcomeMsg)

			// Делаем запрос к сервису меню
			query := fmt.Sprintf("%s/api/v1/menus/getMeal?user_id=%s", b.menuServiceURL, userName)
			log.Println(time.Now().String() + " -> " + query)
			resp, err := http.Get(query)
			if err != nil {
				// b.api.Send(stubMsg)
				b.sendErrorMessage(update.CallbackQuery.Message.Chat.ID, "Ошибка при получении данных")
				continue
			}
			log.Println(time.Now().String() + " -> got responce")
			defer resp.Body.Close()

			// Проверяем статус ответа
			if resp.StatusCode != http.StatusOK {
				// b.api.Send(stubMsg)
				body, _ := io.ReadAll(resp.Body)
				b.sendErrorMessage(update.CallbackQuery.Message.Chat.ID, "Статус ответа не ОК\n"+string(body))
				continue
			}

			// Парсим ответ
			var mealResp MealResponse
			if err := json.NewDecoder(resp.Body).Decode(&mealResp); err != nil {
				// b.api.Send(stubMsg)
				b.sendErrorMessage(update.CallbackQuery.Message.Chat.ID, "Ошибка при разборе данных")
				continue
			}
			log.Println(mealResp.Meal)
			log.Println(mealResp.ShoppingList)

			// Формируем сообщение для пользователя
			message := "🍽 *Следующий прием пищи:*\n\n"
			for i, dish := range mealResp.Meal.DishName {
				message += fmt.Sprintf("🍳 %s\n", dish)
				if i < len(mealResp.Meal.Recipe) {
					var rd RecipeData
					err := json.Unmarshal([]byte(mealResp.Meal.Recipe[i]), &rd)
					if err != nil {
						// Если вдруг не удалось распарсить, покажем сырой вариант, но лучше этого не допускать
						message += fmt.Sprintf("📝 Рецепт: %s\n\n", mealResp.Meal.Recipe[i])
					} else {
						// Форматируем красиво
						message += "📝 Рецепт:\n"
						for _, step := range rd.Steps {
							message += fmt.Sprintf("- %s\n", step)
						}
						message += "\nИнгредиенты:\n"
						for _, ing := range rd.Ingredients {
							message += fmt.Sprintf("- %s: %.0f %s\n", ing.ProductID, ing.Amount, ing.Unit)
						}
						message += "\n"
					}
				}
			}

			if mealResp.ShoppingList != "" {
				message += fmt.Sprintf("\n%s", formatShoppingList(mealResp.ShoppingList))
			}
			log.Println(message)
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, message) // escapeUnderscores(message))
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
	}

	// Отбивка остановки бота
	for chatID, userName := range b.knownUsers {
		id, err := strconv.ParseInt(chatID, 10, 64)
		if err != nil {
			fmt.Println("Ошибка преобразования chatID:", err)
			continue
		}
		message := fmt.Sprintf("🍖 %s, Бот остановлен.\nДа начнется сушка!", userName)
		msg := tgbotapi.NewMessage(id, message)
		b.api.Send(msg)
	}

	return nil
}

func (b *Bot) sendErrorMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("❌ Ошибка: %s", text))

	// Добавляем кнопку для повторной попытки
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Попробовать снова", "get_meal"),
		),
	)
	msg.ReplyMarkup = keyboard

	b.api.Send(msg)
}

// escapeUnderscores — функция, заменяющая "_" на "\_"
func escapeUnderscores(input string) string {
	return strings.ReplaceAll(input, "_", "\\_")
}
