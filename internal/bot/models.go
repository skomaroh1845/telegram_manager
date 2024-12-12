package bot

import "time"

// MealResponse представляет ответ от сервиса меню
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

// Ingredient описывает один ингредиент рецепта
type Ingredient struct {
	Unit      string  `json:"unit"`
	Amount    float64 `json:"amount"`
	ProductID string  `json:"product_id"`
}

// RecipeData описывает структуру рецепта (шаги и ингредиенты)
type RecipeData struct {
	Steps       []string     `json:"steps"`
	Ingredients []Ingredient `json:"ingredients"`
}

// ShoppingListItem описывает один продукт в списке покупок
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

// ShoppingListResponse описывает структуру JSON с продуктами
type ShoppingListResponse struct {
	Products []ShoppingListItem `json:"products"`
}
