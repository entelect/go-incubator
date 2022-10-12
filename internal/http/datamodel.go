package http

type Recipe struct {
	Name        string   `json:"name"`
	Ingredients []string `json:"ingredients"`
}

type Recipes struct {
	Recipes []Recipe `json:"recipes"`
}

// UsesIngredient returns true if the Recipe uses the specified ingredient
func (r *Recipe) UsesIngredient(ingredient string) bool {
	for _, v := range r.Ingredients {
		if v == ingredient {
			return true
		}
	}

	return false
}

// UsesIngredients returns true if the Recipe uses all of the specified ingredients
func (r *Recipe) UsesIngredients(ingredients []string) bool {
	for _, v := range ingredients {
		if !r.UsesIngredient(v) {
			return false
		}
	}

	return true
}
