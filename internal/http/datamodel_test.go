package http

import (
	"testing"
)

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

func TestRecipe_AddIngredient(t *testing.T) {
	type args struct {
		len      int
		contains bool
	}
	tests := []struct {
		name       string
		r          *Recipe
		ingredient string
		want       args
	}{
		{
			name:       "1",
			r:          &Recipe{Name: "Test1", Ingredients: []string{}},
			ingredient: "one",
			want:       args{len: 1, contains: true},
		},
		{
			name:       "2",
			r:          &Recipe{Name: "Test2", Ingredients: []string{"one"}},
			ingredient: "one",
			want:       args{len: 1, contains: true},
		},
		{
			name:       "3",
			r:          &Recipe{Name: "Test3", Ingredients: []string{"one"}},
			ingredient: "two",
			want:       args{len: 2, contains: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r.AddIngredient(tt.ingredient)
			got := args{
				len:      len(tt.r.Ingredients),
				contains: tt.r.UsesIngredient(tt.ingredient),
			}
			if got != tt.want {
				t.Errorf("AddIngredient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRecipe_String(t *testing.T) {
	tests := []struct {
		name string
		r    Recipe
		want string
	}{
		{
			name: "1",
			r:    Recipe{Name: "Test Recipe 1", Ingredients: []string{"One", "Two"}},
			want: "Test Recipe 1\n  - One\n  - Two",
		},
		{
			name: "2",
			r:    Recipe{Ingredients: []string{"One", "Two"}},
			want: "\n  - One\n  - Two",
		},
		{
			name: "3",
			r:    Recipe{Name: "Test Recipe 2"},
			want: "Test Recipe 2",
		},
		{
			name: "4",
			r:    Recipe{},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.String(); got != tt.want {
				t.Errorf("Recipe.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
