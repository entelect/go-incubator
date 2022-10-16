package mysqldb

import (
	"database/sql"
	"fmt"
	"go-incubator/internal/persistence"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MySqlDB struct {
	db *sql.DB
}

func NewMySqlDB(connectionString string) (MySqlDB, error) {
	msdb := MySqlDB{}
	var err error

	msdb.db, err = sql.Open("mysql", connectionString)
	if err != nil {
		return msdb, fmt.Errorf("opening database: %w", err)
	}

	msdb.db.SetConnMaxLifetime(time.Minute * 3)
	msdb.db.SetMaxOpenConns(10)
	msdb.db.SetMaxIdleConns(10)

	err = msdb.db.Ping()
	if err != nil {
		return msdb, fmt.Errorf("pinging database: %w", err)
	}

	return msdb, nil
}

func (mysql *MySqlDB) AddRecipe(recipe persistence.Recipe) error {

	// Start SQL transaction
	tx, err := mysql.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert all ingredients from recipe, ignoring those that are already in db
	for _, ingredient := range recipe.Ingredients {
		_, err := tx.Exec("INSERT IGNORE INTO ingredients (name) VALUES (?)", ingredient)
		if err != nil {
			return fmt.Errorf("writing ingredient: %w", err)
		}
	}

	// Delete existing ingredients relationships for recipe (if it does exist)
	_, err = tx.Exec("DELETE FROM recipe_ingredients WHERE recipe_id = (SELECT id FROM recipes WHERE name = ? LIMIT 1)", recipe.Name)
	if err != nil {
		return fmt.Errorf("adding recipe: %w", err)
	}

	// Insert recipe, ignoring it if it is already in db
	_, err = tx.Exec("INSERT IGNORE INTO recipes (name) VALUES (?)", recipe.Name)
	if err != nil {
		return fmt.Errorf("adding recipe: %w", err)
	}

	// Add ingredient relationships
	for _, ingredient := range recipe.Ingredients {
		_, err := tx.Exec("INSERT INTO recipe_ingredients (recipe_id, ingredient_id) SELECT (SELECT id FROM recipes WHERE name = ? LIMIT 1), id FROM ingredients WHERE name = ?", recipe.Name, ingredient)
		if err != nil {
			return fmt.Errorf("adding ingredient: %w", err)
		}
	}

	// Commit the transaction.
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

func (mysql *MySqlDB) GetRecipe(name string) (persistence.Recipe, error) {
	recipe := persistence.Recipe{Name: name}
	var iname string

	rows, err := mysql.db.Query(`
		SELECT I.name FROM recipes R
		INNER JOIN recipe_ingredients RI ON RI.recipe_id = r.id
		INNER JOIN ingredients I ON I.id = RI.ingredient_id
		WHERE R.name = ?
		ORDER BY I.name`,
		name,
	)
	if err != nil {
		return recipe, fmt.Errorf("executing query :%w", err)
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&iname)
		if err != nil {
			return recipe, fmt.Errorf("reading ingredient name: %w", err)
		}
		recipe.Ingredients = append(recipe.Ingredients, iname)
	}

	if len(recipe.Ingredients) == 0 {
		return recipe, persistence.ErrNoResults
	}

	return recipe, nil
}

func (mysql *MySqlDB) FindRecipes(ingredients []string) ([]persistence.Recipe, error) {
	recipes := []persistence.Recipe{}

	// convert this ingredients to a slice of type any, which is what
	// the func (*sql.DB).Query(query string, args ...any) requires
	var args []any
	for _, ingredient := range ingredients {
		args = append(args, ingredient)
	}
	args = append(args, len(ingredients))

	// Construct a statement which includes all the ingredients we are looking for
	stmt := `
	SELECT R.name FROM recipe_ingredients RI
	INNER JOIN recipes R ON R.id = RI.recipe_id
	INNER JOIN ingredients I ON I.id = RI.ingredient_id
	WHERE I.name IN (?` + strings.Repeat(",?", len(ingredients)-1) + `)
	GROUP BY RI.recipe_id
	HAVING COUNT(*) = ?`
	rows, err := mysql.db.Query(stmt, args...)
	if err != nil {
		return nil, fmt.Errorf("finding recipes: %w", err)
	}

	var rname string
	for rows.Next() {
		err = rows.Scan(&rname)
		if err != nil {
			return nil, fmt.Errorf("reading recipe: %w", err)
		}
		recipe, err := mysql.GetRecipe(rname)
		if err != nil {
			return nil, fmt.Errorf("reading recipe: %w", err)
		}
		recipes = append(recipes, recipe)
	}

	return recipes, nil
}
