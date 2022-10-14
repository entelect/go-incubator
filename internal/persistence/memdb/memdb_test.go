package memdb

import (
	"go-incubator/internal/persistence"
	"reflect"
	"testing"
)

func TestNewMemDB(t *testing.T) {
	tests := []struct {
		name    string
		want    MemDB
		wantErr bool
	}{
		{
			name:    "1",
			want:    MemDB{recipes: make(map[string]persistence.Recipe)},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMemDB()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMemDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMemDB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemDB_AddRecipe(t *testing.T) {
	db, _ := NewMemDB()
	tests := []struct {
		name    string
		db      *MemDB
		recipe  persistence.Recipe
		wantErr bool
		wantLen int
	}{
		{
			name:    "1",
			db:      &db,
			recipe:  persistence.Recipe{Name: "Meatballs", Ingredients: []string{"Ground Beef", "Tomato"}},
			wantErr: false,
			wantLen: 1,
		},
		{
			name:    "2",
			db:      &db,
			recipe:  persistence.Recipe{Name: "Meatballs", Ingredients: []string{"Tomato", "Ground Beef"}},
			wantErr: false,
			wantLen: 1,
		},
		{
			name:    "3",
			db:      &db,
			recipe:  persistence.Recipe{Name: "Cheese Fondue", Ingredients: []string{"Gruyere", "Emmental"}},
			wantErr: false,
			wantLen: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.db.AddRecipe(tt.recipe); (err != nil) != tt.wantErr {
				t.Errorf("MemDB.AddRecipe() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(db.recipes) != tt.wantLen {
				t.Errorf("MemDB.AddRecipe() len = %v, wantLEn %v", len(db.recipes), tt.wantLen)
			}
		})
	}
}

func TestMemDB_GetRecipe(t *testing.T) {
	db, _ := NewMemDB()
	db.recipes["Cheese Fondue"] = persistence.Recipe{Name: "Cheese Fondue", Ingredients: []string{"Gruyere", "Emmental"}}
	db.recipes["Mac & Cheese"] = persistence.Recipe{Name: "Mac & Cheese", Ingredients: []string{"Mozzarella", "Macaroni"}}
	db.recipes["SpagBol"] = persistence.Recipe{Name: "SpagBol", Ingredients: []string{"Spaghetti", "Ground Beef", "Tomato"}}
	db.recipes["BLT"] = persistence.Recipe{Name: "BLT", Ingredients: []string{"Tomato", "Bacon", "Lettuce"}}
	db.recipes["Greek Salad"] = persistence.Recipe{Name: "Greek Salad", Ingredients: []string{"Feta", "Tomato", "Cucumber"}}
	db.recipes["Caprese Salad"] = persistence.Recipe{Name: "Caprese Salad", Ingredients: []string{"Mozzarella", "Tomato"}}
	db.recipes["Meatballs"] = persistence.Recipe{Name: "Meatballs", Ingredients: []string{"Ground Beef", "Tomato"}}

	tests := []struct {
		name    string
		db      *MemDB
		rname   string
		want    persistence.Recipe
		wantErr bool
	}{
		{
			name:    "1",
			db:      &db,
			rname:   "Cheese Fondue",
			want:    persistence.Recipe{Name: "Cheese Fondue", Ingredients: []string{"Gruyere", "Emmental"}},
			wantErr: false,
		},
		{
			name:    "2",
			db:      &db,
			rname:   "SpagBol",
			want:    persistence.Recipe{Name: "SpagBol", Ingredients: []string{"Spaghetti", "Ground Beef", "Tomato"}},
			wantErr: false,
		},
		{
			name:    "3",
			db:      &db,
			rname:   "Pizza",
			want:    persistence.Recipe{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.db.GetRecipe(tt.rname)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemDB.GetRecipe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MemDB.GetRecipe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemDB_FindRecipes(t *testing.T) {
	db, _ := NewMemDB()
	db.recipes["Cheese Fondue"] = persistence.Recipe{Name: "Cheese Fondue", Ingredients: []string{"Gruyere", "Emmental"}}
	db.recipes["Mac & Cheese"] = persistence.Recipe{Name: "Mac & Cheese", Ingredients: []string{"Mozzarella", "Macaroni"}}
	db.recipes["SpagBol"] = persistence.Recipe{Name: "SpagBol", Ingredients: []string{"Spaghetti", "Ground Beef", "Tomato"}}
	db.recipes["BLT"] = persistence.Recipe{Name: "BLT", Ingredients: []string{"Tomato", "Bacon", "Lettuce"}}
	db.recipes["Greek Salad"] = persistence.Recipe{Name: "Greek Salad", Ingredients: []string{"Feta", "Tomato", "Cucumber"}}
	db.recipes["Caprese Salad"] = persistence.Recipe{Name: "Caprese Salad", Ingredients: []string{"Mozzarella", "Tomato"}}
	db.recipes["Meatballs"] = persistence.Recipe{Name: "Meatballs", Ingredients: []string{"Ground Beef", "Tomato"}}

	tests := []struct {
		name        string
		db          *MemDB
		ingredients []string
		want        []persistence.Recipe
		wantErr     bool
	}{
		{
			name:        "1",
			db:          &db,
			ingredients: []string{"Gruyere", "Emmental"},
			want:        []persistence.Recipe{{Name: "Cheese Fondue", Ingredients: []string{"Gruyere", "Emmental"}}},
		},
		{
			name:        "2",
			db:          &db,
			ingredients: []string{"Emmental", "Gruyere"},
			want:        []persistence.Recipe{{Name: "Cheese Fondue", Ingredients: []string{"Gruyere", "Emmental"}}},
		},
		{
			name:        "3",
			db:          &db,
			ingredients: []string{"Tomato"},
			want:        []persistence.Recipe{{Name: "BLT", Ingredients: []string{"Tomato", "Bacon", "Lettuce"}}, {Name: "Caprese Salad", Ingredients: []string{"Mozzarella", "Tomato"}}, {Name: "Greek Salad", Ingredients: []string{"Feta", "Tomato", "Cucumber"}}, {Name: "Meatballs", Ingredients: []string{"Ground Beef", "Tomato"}}, {Name: "SpagBol", Ingredients: []string{"Spaghetti", "Ground Beef", "Tomato"}}},
		},
		{
			name:        "4",
			db:          &db,
			ingredients: []string{"Tomato", "Onion"},
			want:        []persistence.Recipe{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.db.FindRecipes(tt.ingredients)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemDB.FindRecipes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MemDB.FindRecipes() = %v, want %v", got, tt.want)
			}
		})
	}
}
