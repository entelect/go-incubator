package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type HttpClient struct {
	client  *http.Client
	address string
	apiKey  string
}

// NewHttpClient creates and returns a new HttpClient pointing at the specified address on the specified port
func NewHttpClient(baseUrl string, port int, apiKey string) (HttpClient, error) {
	c := HttpClient{
		client:  &http.Client{},
		address: fmt.Sprintf("http://%s:%d", baseUrl, port),
		apiKey:  apiKey,
	}

	return c, nil
}

// AddRecipe calls the `POST /recipe` endpoint
func (c *HttpClient) AddRecipe(recipe Recipe) error {
	payload, err := json.Marshal(recipe)
	if err != nil {
		return fmt.Errorf("marshalling recipe: %w", err)
	}

	postBody := bytes.NewBuffer(payload)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/recipe", c.address), postBody)
	if err != nil {
		return fmt.Errorf("creating http request: %w", err)
	}
	req.Header.Add("X-Api-Key", c.apiKey)

	res, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("calling http endpoint: %w", err)
	}
	defer res.Body.Close()

	// Even if we're not interested in the body, we have to read all of
	// it in order for the htp.Client to be reusable
	_, err = io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf(res.Status)
	}

	return nil
}

// GetRecipe calls the `GET /recipe/{name}` endpoint
func (c *HttpClient) GetRecipe(name string) (*Recipe, error) {
	var recipe *Recipe
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/recipe/%s", c.address, url.QueryEscape(name)), nil)
	if err != nil {
		return nil, fmt.Errorf("creating http request: %w", err)
	}
	req.Header.Add("X-Api-Key", c.apiKey)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling http endpoint: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(res.Status)
	}

	err = json.Unmarshal(body, &recipe)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling response: %v", err)
	}

	return recipe, nil
}

// SearchByIngredients calls the `GET /recipes?ingredients={list of ingredients}` endpoint
func (c *HttpClient) SearchByIngredients(ingredients []string) ([]Recipe, error) {
	var recipes Recipes
	list := strings.Join(ingredients, ",")
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/recipes?ingredients=%s", c.address, url.QueryEscape(list)), nil)
	if err != nil {
		return nil, fmt.Errorf("creating http request: %w", err)
	}
	req.Header.Add("X-Api-Key", c.apiKey)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling http endpoint: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(res.Status)
	}

	err = json.Unmarshal(body, &recipes)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling response: %v", err)
	}

	return recipes.Recipes, nil
}

func (c *HttpClient) Benchmarks(duration time.Duration) {
	numRoutines := 100
	fmt.Printf("Calling SearchByIngredients([]string{\"Tomato\"}) on %d concurrent routines for %s, please wait\n", numRoutines, duration)
	fmt.Println("┎──────────────────────────────────────────────────────────┒")

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
