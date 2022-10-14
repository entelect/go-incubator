package persistence

import (
	"testing"
)

func TestRecipe_UsesIngredient(t *testing.T) {
	recipe := Recipe{
		Name:        "Test Name",
		Ingredients: []string{"one", "two", "three"},
	}
	tests := []struct {
		name       string
		r          *Recipe
		ingredient string
		want       bool
	}{
		{name: "1", r: &recipe, ingredient: "one", want: true},
		{name: "2", r: &recipe, ingredient: "two", want: true},
		{name: "3", r: &recipe, ingredient: "three", want: true},
		{name: "4", r: &recipe, ingredient: "four", want: false},
		{name: "5", r: &recipe, ingredient: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.UsesIngredient(tt.ingredient); got != tt.want {
				t.Errorf("Recipe.UsesIngredient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUsesIngredient(t *testing.T) {
	recipe := Recipe{
		Name:        "Test Name",
		Ingredients: []string{"one", "two", "three"},
	}

	tests := []struct {
		ingredient string
		want       bool
	}{
		{"one", true},
		{"two", true},
		{"three", true},
		{"four", false},
		{"", false},
	}

	for i, test := range tests {
		got := recipe.UsesIngredient(test.ingredient)
		if got != test.want {
			t.Errorf(`test %d (ingredient="%s"), got %t, want %t`, i+1, test.ingredient, got, test.want)
		}
	}
}

func TestUsesIngredients(t *testing.T) {
	recipe := Recipe{
		Name:        "Test Name",
		Ingredients: []string{"one", "two", "three"},
	}

	tests := []struct {
		ingredients []string
		want        bool
	}{
		{[]string{"one"}, true},
		{[]string{"one", "two"}, true},
		{[]string{"three", "two", "one"}, true},
		{[]string{"four"}, false},
		{[]string{"one", "two", "three", "four"}, false},
		{[]string{}, true},
	}

	for i, test := range tests {
		got := recipe.UsesIngredients(test.ingredients)
		if got != test.want {
			t.Errorf(`test %d (ingredient="%s"), got %t, want %t`, i+1, test.ingredients, got, test.want)
		}
	}
}

func TestRecipe_UsesIngredients(t *testing.T) {
	recipe := Recipe{
		Name:        "Test Name",
		Ingredients: []string{"one", "two", "three"},
	}

	tests := []struct {
		name        string
		r           *Recipe
		ingredients []string
		want        bool
	}{
		{name: "1", r: &recipe, ingredients: []string{"one"}, want: true},
		{name: "2", r: &recipe, ingredients: []string{"one", "two"}, want: true},
		{name: "3", r: &recipe, ingredients: []string{"three", "two", "one"}, want: true},
		{name: "4", r: &recipe, ingredients: []string{"four"}, want: false},
		{name: "5", r: &recipe, ingredients: []string{"one", "two", "three", "four"}, want: false},
		{name: "6", r: &recipe, ingredients: []string{}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.UsesIngredients(tt.ingredients); got != tt.want {
				t.Errorf("Recipe.UsesIngredients() = %v, want %v", got, tt.want)
			}
		})
	}
}
