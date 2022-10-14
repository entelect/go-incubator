package persistence

import "errors"

type Recipe struct {
	Name        string
	Ingredients []string
}

// ErrNoResults is returned when no results are found
var ErrNoResults = errors.New("datastore: no results found")

// Persistence is an interface that can be implemented by database structures
type Persistence interface {
	AddRecipe(Recipe) error
	GetRecipe(string) (Recipe, error)
	FindRecipes([]string) ([]Recipe, error)
}
