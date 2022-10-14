package memdb

import (
	"go-incubator/internal/persistence"
	"sort"
)

type MemDB struct {
	recipes map[string]persistence.Recipe
}

func NewMemDB() (MemDB, error) {
	db := MemDB{
		recipes: make(map[string]persistence.Recipe),
	}

	return db, nil
}

func (db *MemDB) AddRecipe(recipe persistence.Recipe) error {
	db.recipes[recipe.Name] = recipe

	return nil
}

func (db *MemDB) GetRecipe(name string) (persistence.Recipe, error) {
	recipe, ok := db.recipes[name]
	if !ok {
		return recipe, persistence.ErrNoResults
	}

	return recipe, nil
}

func (db *MemDB) FindRecipes(ingredients []string) ([]persistence.Recipe, error) {
	// We want to return the list of recipes in alphabetical order (by name)
	// To do that, we first extract map keys into a slice, then sort the slice,
	// then iterate over the slice, to obtain map entries in alphabetical order
	keys := make([]string, 0, len(db.recipes))
	for k := range db.recipes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	recipes := make([]persistence.Recipe, 0)
	for _, k := range keys {
		recipe := db.recipes[k]
		if recipe.UsesIngredients(ingredients) {
			recipes = append(recipes, recipe)
		}
	}

	return recipes, nil
}
