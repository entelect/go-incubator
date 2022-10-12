package http

import "testing"

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
