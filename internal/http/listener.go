package http

import (
	"context"
	"encoding/json"
	"fmt"
	"go-incubator/internal/persistence"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type HttpServer struct {
	server *http.Server
	port   int
	apiKey string
	db     persistence.Persistence
}

// NewHttpServer creates and returns a new HttpServer with a listener on the specified port
func NewHttpServer(port int, apiKey string, persistence persistence.Persistence) (HttpServer, error) {
	s := HttpServer{server: &http.Server{Addr: fmt.Sprintf(":%d", port)},
		port:   port,
		apiKey: apiKey,
		db:     persistence,
	}

	mux := http.NewServeMux()
	mux.Handle("/", s.stdHeaders(s.auth(s.tracer(http.HandlerFunc(s.router)))))
	s.server.Handler = mux

	return s, nil
}

// Start initiates the HTTP listener of the received HttpServer
func (s *HttpServer) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		fmt.Printf("starting HTTP listener on port %d\n", s.port)
		defer fmt.Printf("HTP listener on port %d stopped\n", s.port)
		if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Printf("http server error: %v\n", err)
		}
	}()
}

// Stop terminates the HTTP listener of the received HttpServer
func (s *HttpServer) Stop() {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := s.server.Shutdown(ctxTimeout); err != nil {
		panic(err) // failure/timeout shutting down the server gracefully
	}
}

// router is the top level Handler which directs incoming traffic to the appropriate endpoint Handlers
func (s *HttpServer) router(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" && r.RequestURI == "/recipe" {
		s.addRecipe(w, r)
		return
	}

	if r.Method == "GET" && strings.HasPrefix(r.RequestURI, "/recipe/") {
		s.getRecipe(w, r)
		return
	}

	if r.Method == "GET" && strings.HasPrefix(r.RequestURI, "/recipes") {
		s.findRecipes(w, r)
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

// addRecipe is the Handler for adding new recipes
func (s *HttpServer) addRecipe(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)

	recipe := Recipe{}
	err := json.Unmarshal(body, &recipe)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error unmarshalling recipe"))
		return
	}

	if recipe.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("no name specified"))
		return
	}

	if len(recipe.Ingredients) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("no ingredients specified"))
		return
	}

	err = s.db.AddRecipe(persistence.Recipe(recipe))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error writing recipe to database"))
	}
}

// getRecipe is the Handler for retrieving a recipe by name
func (s *HttpServer) getRecipe(w http.ResponseWriter, r *http.Request) {
	name, err := url.QueryUnescape(strings.TrimPrefix(r.RequestURI, "/recipe/"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	recipe, err := s.db.GetRecipe(name)
	if err == persistence.ErrNoResults {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error reading recipe from database"))
		return
	}

	rsp, err := json.Marshal(Recipe(recipe))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error marshalling recipe into json"))
		return
	}

	w.Write(rsp)
}

// findRecipes is the Handler for listing recipes by ingredients
func (s *HttpServer) findRecipes(w http.ResponseWriter, r *http.Request) {
	unescaped, err := url.QueryUnescape(strings.TrimPrefix(r.RequestURI, "/recipes"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	elems := strings.Split(unescaped, "?")
	if len(elems) > 2 {
		w.WriteHeader((http.StatusBadRequest))
		return
	}

	var params []string
	if len(elems) > 1 {
		params = strings.Split(elems[1], "&")
	}

	var ingredients []string
	for _, v := range params {
		if strings.HasPrefix(v, "ingredients=") {
			ingredients = strings.Split(strings.TrimPrefix(v, "ingredients="), ",")
		}
	}

	if len(ingredients) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("no ingredients specified"))
		return
	}

	dbrecipes, err := s.db.FindRecipes(ingredients)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error reading recipes from database"))
		return
	}

	// Convert []persistence.Recipe to []Recipe
	recipes := []Recipe{}
	for _, r := range dbrecipes {
		recipes = append(recipes, Recipe(r))
	}

	rsp, err := json.Marshal(Recipes{Recipes: recipes})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error marshalling recipes into json"))
		return
	}

	w.Write(rsp)
}

// tracer measures the time it took for each API call to be processed
func (s *HttpServer) tracer(originalHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		endpoint := fmt.Sprintf("%s %s", r.Method, r.RequestURI)
		start := time.Now()
		originalHandler.ServeHTTP(w, r)
		end := time.Now()
		fmt.Println(endpoint, end.Sub(start))
	})
}

// auth checks that API requests contain required API key
func (s *HttpServer) auth(originalHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		keys := r.Header["X-Api-Key"]
		keyfound := false
		for _, v := range keys {
			if v == s.apiKey {
				keyfound = true
			}
		}
		if keyfound {
			originalHandler.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	})
}

// stdHeaders adds some standard headers into HTTP responses
func (s *HttpServer) stdHeaders(originalHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		originalHandler.ServeHTTP(w, r)
	})
}
