package main

import (
	"fmt"
	config "go-incubator/internal/configuration"
	"go-incubator/internal/grpc"
	"go-incubator/internal/helpers"
	"go-incubator/internal/http"
	"go-incubator/internal/ui"
	"time"
)

func main() {
	cfg, err := config.ReadConfig("INCUBATOR_")
	if err != nil {
		fmt.Printf("error reading config: %v\n", err)
		return
	}

	grpcClient, err := grpc.NewGrpcClient(cfg.Address, cfg.GrpcPort, cfg.ApiKey)
	if err != nil {
		fmt.Printf("error creating gRPC client: %v\n", err)
		return
	}

	for {
		action := ui.Selection("What would you like to do?", []string{"Add a recipe", "Get a recipe", "Search by ingredients", "Run Benchmarks", "Quit"})
		fmt.Println()

		switch action {
		case "Add a recipe":
			fmt.Println("Adding a recipe:")
			newRecipe := http.Recipe{}
			newRecipe.Name = ui.GetValue("Enter name of recipe -> ")
			addIngredients := true
			for addIngredients {
				ingredient := ui.GetValue("Enter ingredient name (blank to stop) -> ")
				if ingredient == "" {
					addIngredients = false
				} else {
					newRecipe.AddIngredient(ingredient)
				}
			}
			fmt.Println()
			err = grpcClient.AddRecipe(newRecipe)
			if err != nil {
				fmt.Printf("Something went wrong when we tried to add the new recipe: %v\n", err)
			} else {
				fmt.Printf("%s added successfully\n", newRecipe.Name)
			}
		case "Get a recipe":
			fmt.Println("Getting a recipe:")
			name := ui.GetValue("Enter name of recipe -> ")
			fmt.Println()

			recipe, err := grpcClient.GetRecipe(name)
			if err != nil {
				fmt.Printf("Something went wrong when we tried to get the recipe: %v\n", err)
			} else {
				if recipe == nil {
					fmt.Printf("Sorry, no recipe for %s found\n", name)
				} else {
					fmt.Printf("Recipe found:\n%s\n", recipe)
				}
			}
		case "Search by ingredients":
			fmt.Println("Finding a recipe by ingredients:")
			ingredients := []string{}
			addIngredients := true
			for addIngredients {
				ingredient := ui.GetValue("Enter ingredient name (blank to stop) -> ")
				if ingredient == "" {
					addIngredients = false
				} else {
					if !helpers.StringSliceContains(ingredients, ingredient) {
						ingredients = append(ingredients, ingredient)
					}
				}
			}
			fmt.Println()
			fmt.Printf("Searching for recipes that make use of %+v\n", ingredients)

			recipes, err := grpcClient.SearchByIngredients(ingredients)
			if err != nil {
				fmt.Printf("Something went wrong when we tried to find the recipes: %v\n", err)
			} else {
				if len(recipes) == 0 {
					fmt.Printf("Sorry, no recipes found using these ingredients\n")
				} else {
					fmt.Printf("Found the following %d recipes using these ingredients:\n\n", len(recipes))
					for _, r := range recipes {
						fmt.Println(r)
						fmt.Println()
					}
				}
			}
		case "Run Benchmarks":
			grpcClient.Benchmarks(1 * time.Minute)
		case "Quit":
			fmt.Println("OK bye")
			return
		}
		fmt.Println()
	}
}
