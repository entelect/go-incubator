package http

import "fmt"

type Recipe struct {
	Name        string   `json:"name"`
	Ingredients []string `json:"ingredients"`
}

type Recipes struct {
	Recipes []Recipe `json:"recipes"`
}

func (r Recipe) String() string {
	rsp := r.Name
	for _, v := range r.Ingredients {
		rsp = fmt.Sprintf("%s\n  - %s", rsp, v)
	}

	return rsp
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

// AddIngredient adds an ingredient to the Recipe only if it isn't already listed
func (r *Recipe) AddIngredient(ingredient string) {
	if !r.UsesIngredient(ingredient) {
		r.Ingredients = append(r.Ingredients, ingredient)
	}
}
