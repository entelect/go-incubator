package persistence

import "errors"

type Recipe struct {
	Name        string
	Ingredients []string
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

// ErrNoResults is returned when no results are found
var ErrNoResults = errors.New("datastore: no results found")

// Persistence is an interface that can be implemented by database structures
type Persistence interface {
	AddRecipe(Recipe) error
	GetRecipe(string) (Recipe, error)
	FindRecipes([]string) ([]Recipe, error)
}
