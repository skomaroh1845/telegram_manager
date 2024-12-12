package bot

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// formatShoppingList форматирует список покупок в читаемый вид
func formatShoppingList(jsonStr string) string {
	var response ShoppingListResponse
	if err := json.Unmarshal([]byte(jsonStr), &response); err != nil {
		log.Println(fmt.Errorf("failed to unmarshal shopping list: %w", err))
		return "❌ Ошибка в формате списка покупок"
	}

	var result strings.Builder
	result.WriteString("🛒 *Список покупок:*\n")

	for _, item := range response.Products {
		itemName := item.Name
		if itemName == "" {
			itemName = item.ID
		}

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

// buildMealMessage формирует сообщение о следующем приеме пищи на основе MealResponse
func buildMealMessage(mealResp *MealResponse) string {
	var message strings.Builder
	message.WriteString("🍽 *Следующий прием пищи:*\n\n")

	for i, dish := range mealResp.Meal.DishName {
		message.WriteString(fmt.Sprintf("🍳 %s\n", dish))
		if i < len(mealResp.Meal.Recipe) {
			var rd RecipeData
			err := json.Unmarshal([]byte(mealResp.Meal.Recipe[i]), &rd)
			if err != nil {
				// Если вдруг не удалось распарсить, показываем сырой вариант
				message.WriteString(fmt.Sprintf("📝 Рецепт: %s\n\n", mealResp.Meal.Recipe[i]))
			} else {
				// Форматируем рецепт
				message.WriteString("📝 Рецепт:\n")
				for _, step := range rd.Steps {
					message.WriteString(fmt.Sprintf("- %s\n", step))
				}
				message.WriteString("\nИнгредиенты:\n")
				for _, ing := range rd.Ingredients {
					message.WriteString(fmt.Sprintf("- %s: %.0f %s\n", ing.ProductID, ing.Amount, ing.Unit))
				}
				message.WriteString("\n")
			}
		}
	}

	if mealResp.ShoppingList != "" {
		message.WriteString("\n")
		message.WriteString(formatShoppingList(mealResp.ShoppingList))
	}

	return message.String()
}

// escapeUnderscores экранирует символы подчеркивания для Markdown
// func escapeUnderscores(input string) string {
// 	return strings.ReplaceAll(input, "_", "\\_")
// }
