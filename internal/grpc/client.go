package grpc

import (
	"context"
	"fmt"
	"go-incubator/internal/http"
	"go-incubator/proto"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcMetadata "google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type GrpcClient struct {
	client proto.RecipeServiceClient
	apiKey string
}

// NewGrpcClient creates and returns a new GrpcClient pointing at the specified address on the specified port
func NewGrpcClient(baseUrl string, port int, apiKey string) (GrpcClient, error) {
	c := GrpcClient{
		apiKey: apiKey,
	}

	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", baseUrl, port),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(c.auth),
	)
	if err != nil {
		return c, fmt.Errorf("creating gRPC client: %w", err)
	}

	c.client = proto.NewRecipeServiceClient(conn)

	return c, nil
}

// auth is an interceptor function that adds the API Key to the outgoing context
func (c *GrpcClient) auth(
	ctx context.Context,
	method string,
	req interface{},
	reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	return invoker(
		grpcMetadata.AppendToOutgoingContext(ctx, "x-api-key", c.apiKey),
		method, req, reply, cc, opts...,
	)
}

// AddRecipe calls the `RecipeService/AddRecipe` gRPC function
func (c *GrpcClient) AddRecipe(recipe http.Recipe) error {
	_, err := c.client.AddRecipe(
		context.Background(),
		&proto.Recipe{
			Name:        recipe.Name,
			Ingredients: recipe.Ingredients,
		},
	)
	if err != nil {
		return fmt.Errorf("calling gRPC function: %w", err)
	}

	return nil
}

// GetRecipe calls the `RecipeService/GetRecipe` gRPC function
func (c *GrpcClient) GetRecipe(name string) (*http.Recipe, error) {
	rsp, err := c.client.GetRecipe(
		context.Background(),
		&proto.RecipeRequest{Name: name},
	)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.NotFound {
				return nil, nil
			}
		}
		return nil, fmt.Errorf("calling gRPC function: %w", err)
	}

	return &http.Recipe{
		Name:        rsp.Name,
		Ingredients: rsp.Ingredients,
	}, nil
}

// SearchByIngredients calls the `RecipeService/FindRecipes` gRPC function
func (c *GrpcClient) SearchByIngredients(ingredients []string) ([]http.Recipe, error) {
	var recipes []http.Recipe

	rsp, err := c.client.FindRecipes(
		context.Background(),
		&proto.FindRequest{Ingredients: ingredients},
	)
	if err != nil {
		return nil, fmt.Errorf("calling gRPC function: %w", err)
	}

	// Convert *proto.Recipes to []http.Recipe
	for _, r := range rsp.Recipes {
		recipes = append(recipes, http.Recipe{Name: r.Name, Ingredients: r.Ingredients})
	}

	return recipes, nil
}

func (c *GrpcClient) Benchmarks(duration time.Duration) {
	numRoutines := 100
	fmt.Printf("Calling SearchByIngredients([]string{\"Tomato\"}) on %d concurrent routines for %s, please wait\n", numRoutines, duration)
	fmt.Println("????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????")

	done := make(chan interface{})
	cntchan := make(chan int, numRoutines)
	counter := 0
	var wg sync.WaitGroup
	wg.Add(numRoutines)
	defer wg.Wait()

	for i := 0; i < numRoutines; i++ {
		go func(cnt chan int) {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					c.SearchByIngredients([]string{"Tomato"})
					cnt <- 1
				}
			}
		}(cntchan)
	}

	go func(c chan int) {
		for i := range c {
			counter = counter + i
		}
	}(cntchan)

	go func() {
		for {
			select {
			case <-done:
				return
			default:
				fmt.Print(".")
				time.Sleep(1 * time.Second)
			}
		}
	}()

	time.Sleep(duration)
	close(done)

	fmt.Printf("\nSearchByIngredients([]string{\"Tomato\"}) called %d times in %s\n\n", counter, duration)
}
