package http

import (
	"bytes"
	"fmt"
	"go-incubator/internal/persistence"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

type mockdb struct {
	recipes map[string]persistence.Recipe
}

func NewMockDB() *mockdb {
	mdb := &mockdb{}
	mdb.recipes = make(map[string]persistence.Recipe)
	mdb.recipes["Cheese Fondue"] = persistence.Recipe{Name: "Cheese Fondue", Ingredients: []string{"Gruyere", "Emmental"}}
	mdb.recipes["Mac & Cheese"] = persistence.Recipe{Name: "Mac & Cheese", Ingredients: []string{"Mozzarella", "Macaroni"}}
	mdb.recipes["SpagBol"] = persistence.Recipe{Name: "SpagBol", Ingredients: []string{"Spaghetti", "Ground Beef", "Tomato"}}
	mdb.recipes["BLT"] = persistence.Recipe{Name: "BLT", Ingredients: []string{"Tomato", "Bacon", "Lettuce"}}
	mdb.recipes["Greek Salad"] = persistence.Recipe{Name: "Greek Salad", Ingredients: []string{"Feta", "Tomato", "Cucumber"}}
	mdb.recipes["Caprese Salad"] = persistence.Recipe{Name: "Caprese Salad", Ingredients: []string{"Mozzarella", "Tomato"}}
	mdb.recipes["Meatballs"] = persistence.Recipe{Name: "Meatballs", Ingredients: []string{"Ground Beef", "Tomato"}}

	return mdb
}

func (db *mockdb) AddRecipe(recipe persistence.Recipe) error {
	if recipe.Name == "DB Error" {
		return fmt.Errorf("Database Error")
	}
	return nil
}

func (db *mockdb) GetRecipe(name string) (persistence.Recipe, error) {
	if name == "DBError" {
		return persistence.Recipe{}, fmt.Errorf("Database Error")
	}
	r, ok := db.recipes[name]
	if !ok {
		return persistence.Recipe{}, persistence.ErrNoResults
	}

	return r, nil
}

func (db *mockdb) FindRecipes(ingredients []string) ([]persistence.Recipe, error) {
	if strings.Join(ingredients, "") == "DBError" {
		return nil, fmt.Errorf("Database Error")
	}

	keys := make([]string, 0, len(db.recipes))
	for k := range db.recipes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	recipes := make([]persistence.Recipe, 0)
	for _, k := range keys {
		recipe := db.recipes[k]
		if UsesIngredients(recipe, ingredients) {
			recipes = append(recipes, recipe)
		}
	}

	return recipes, nil
}

func UsesIngredient(r persistence.Recipe, ingredient string) bool {
	for _, v := range r.Ingredients {
		if v == ingredient {
			return true
		}
	}

	return false
}

func UsesIngredients(r persistence.Recipe, ingredients []string) bool {
	for _, v := range ingredients {
		if !UsesIngredient(r, v) {
			return false
		}
	}

	return true
}

func captureOutput(f func()) string {
	reader, writer, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	stdout := os.Stdout
	stderr := os.Stderr
	defer func() {
		os.Stdout = stdout
		os.Stderr = stderr
		log.SetOutput(os.Stderr)
	}()
	os.Stdout = writer
	os.Stderr = writer
	log.SetOutput(writer)
	out := make(chan string)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		var buf bytes.Buffer
		wg.Done()
		io.Copy(&buf, reader)
		out <- buf.String()
	}()
	wg.Wait()
	f()
	writer.Close()
	return <-out
}

func TestNewHttpServer(t *testing.T) {
	type args struct {
		port   int
		apiKey string
	}
	tests := []struct {
		name    string
		args    args
		want    HttpServer
		wantErr bool
	}{
		{
			name: "1",
			args: args{port: 1234, apiKey: "1234"},
			want: HttpServer{
				server: &http.Server{
					Addr: fmt.Sprintf(":%d", 1234),
				},
				apiKey: "1234",
				db:     NewMockDB(),
			},
		},
		{
			name: "2",
			args: args{port: 1234, apiKey: ""},
			want: HttpServer{
				server: &http.Server{
					Addr: fmt.Sprintf(":%d", 1234),
				},
				apiKey: "",
				db:     NewMockDB(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewHttpServer(tt.args.port, tt.args.apiKey, tt.want.db)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHttpServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.server.Addr != tt.want.server.Addr || got.apiKey != tt.want.apiKey || !reflect.DeepEqual(got.db, tt.want.db) {
				t.Errorf("NewHttpServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHttpServer_addRecipe(t *testing.T) {
	server, _ := NewHttpServer(1234, "1234", NewMockDB())

	type response struct {
		code int
		body string
	}

	tests := []struct {
		name string
		body string
		want response
	}{
		{
			name: "1",
			body: `{"name":"Meatballs","ingredients":["Ground Beef","Tomato"]}`,
			want: response{
				code: http.StatusOK,
				body: ``,
			},
		},
		{
			name: "2",
			body: `{"ingredients":["Mozzarella","Macaroni"]}`,
			want: response{
				code: http.StatusBadRequest,
				body: `no name specified`,
			},
		},
		{
			name: "3",
			body: `{"name":"Pizza"}`,
			want: response{
				code: http.StatusBadRequest,
				body: `no ingredients specified`,
			},
		},
		{
			name: "4",
			body: `{"name":"Meatballs","ingredients":["Ground Beef,"Tomato"]}`,
			want: response{
				code: http.StatusBadRequest,
				body: `error unmarshalling recipe`,
			},
		},
		{
			name: "5",
			body: `{"name":"DB Error","ingredients":["Ground Beef","Tomato"]}`,
			want: response{
				code: http.StatusInternalServerError,
				body: `error writing recipe to database`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/recipe", strings.NewReader(tt.body))
			server.addRecipe(w, r)

			if w.Code != tt.want.code || w.Body.String() != tt.want.body {
				t.Errorf("addRecipe() = %v, want %v", response{code: w.Code, body: w.Body.String()}, tt.want)
			}
		})
	}
}

func TestHttpServer_getRecipe(t *testing.T) {
	server, _ := NewHttpServer(1234, "1234", NewMockDB())

	type response struct {
		code int
		body string
	}

	tests := []struct {
		name string
		path string
		body string
		want response
	}{
		{
			name: "1",
			path: "/recipe/Cheese%20Fondue",
			want: response{
				code: http.StatusOK,
				body: `{"name":"Cheese Fondue","ingredients":["Gruyere","Emmental"]}`,
			},
		},
		{
			name: "2",
			path: "/recipe/SpagBol",
			want: response{
				code: http.StatusOK,
				body: `{"name":"SpagBol","ingredients":["Spaghetti","Ground Beef","Tomato"]}`,
			},
		},
		{
			name: "3",
			path: "/recipe/Pizza",
			want: response{
				code: http.StatusNotFound,
			},
		},
		{
			name: "4",
			path: "/recipe/DBError",
			want: response{
				code: http.StatusInternalServerError,
				body: "error reading recipe from database",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", tt.path, strings.NewReader(tt.body))
			server.getRecipe(w, r)

			if w.Code != tt.want.code || w.Body.String() != tt.want.body {
				t.Errorf("getRecipe() = %v, want %v", response{code: w.Code, body: w.Body.String()}, tt.want)
			}
		})
	}
}

func TestHttpServer_findRecipes(t *testing.T) {
	server, _ := NewHttpServer(1234, "1234", NewMockDB())

	type response struct {
		code int
		body string
	}

	tests := []struct {
		name string
		path string
		body string
		want response
	}{
		{
			name: "1",
			path: "/recipes?ingredients=Gruyere,Emmental",
			want: response{
				code: http.StatusOK,
				body: `{"recipes":[{"name":"Cheese Fondue","ingredients":["Gruyere","Emmental"]}]}`,
			},
		},
		{
			name: "2",
			path: "/recipes?ingredients=Emmental,Gruyere",
			want: response{
				code: http.StatusOK,
				body: `{"recipes":[{"name":"Cheese Fondue","ingredients":["Gruyere","Emmental"]}]}`,
			},
		},
		{
			name: "3",
			path: "/recipes?ingredients=Tomato",
			want: response{
				code: http.StatusOK,
				body: `{"recipes":[{"name":"BLT","ingredients":["Tomato","Bacon","Lettuce"]},{"name":"Caprese Salad","ingredients":["Mozzarella","Tomato"]},{"name":"Greek Salad","ingredients":["Feta","Tomato","Cucumber"]},{"name":"Meatballs","ingredients":["Ground Beef","Tomato"]},{"name":"SpagBol","ingredients":["Spaghetti","Ground Beef","Tomato"]}]}`,
			},
		},
		{
			name: "4",
			path: "/recipes?ingredients=Tomato,Onion",
			want: response{
				code: http.StatusOK,
				body: `{"recipes":[]}`,
			},
		},
		{
			name: "5",
			path: "/recipes",
			want: response{
				code: http.StatusBadRequest,
				body: `no ingredients specified`,
			},
		},
		{
			name: "6",
			path: "/recipes?ingredients=Tomato?Ingredients=Onion",
			want: response{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "7",
			path: "/recipes?ingredients=DBError",
			want: response{
				code: http.StatusInternalServerError,
				body: "error reading recipes from database",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", tt.path, strings.NewReader(tt.body))
			server.findRecipes(w, r)

			if w.Code != tt.want.code || w.Body.String() != tt.want.body {
				t.Errorf("findRecipes() = %v, want %v", response{code: w.Code, body: w.Body.String()}, tt.want)
			}
		})
	}
}

func TestHttpServer_tracer(t *testing.T) {
	server, _ := NewHttpServer(1234, "1234", NewMockDB())

	type args struct {
		originalHandler http.Handler
	}
	tests := []struct {
		name      string
		s         *HttpServer
		args      args
		wantRegex string
	}{
		{
			name:      "1",
			s:         &server,
			args:      args{originalHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})},
			wantRegex: "GET /tracertest 0s",
		},
		{
			name:      "2",
			s:         &server,
			args:      args{originalHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { time.Sleep(5 * time.Second) })},
			wantRegex: `GET /tracertest 5(.\d+)?s`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/tracertest", nil)

			got := captureOutput(func() { server.tracer(tt.args.originalHandler).ServeHTTP(w, r) })

			match, err := regexp.MatchString(tt.wantRegex, got)
			if err != nil {
				t.Errorf("HttpServer.tracer(): %v", err)
			}
			if !match {
				t.Errorf("HttpServer.tracer() = %v, want %v", got, tt.wantRegex)
			}
		})
	}
}

func TestHttpServer_auth(t *testing.T) {
	server, _ := NewHttpServer(1234, "1234", NewMockDB())

	type response struct {
		code int
		body string
	}

	tests := []struct {
		name string
		key  string
		want response
	}{
		{
			name: "1",
			key:  "1234",
			want: response{
				code: http.StatusOK,
			},
		},
		{
			name: "2",
			want: response{
				code: http.StatusUnauthorized,
			},
		},
		{
			name: "3",
			key:  "4321",
			want: response{
				code: http.StatusUnauthorized,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/authtest", nil)
			if tt.key != "" {
				r.Header.Set("X-Api-Key", tt.key)
			}
			server.auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).ServeHTTP(w, r)

			if w.Code != tt.want.code || w.Body.String() != tt.want.body {
				t.Errorf("auth() = %v, want %v", response{code: w.Code, body: w.Body.String()}, tt.want)
			}
		})
	}
}

func TestHttpServer_stdHeaders(t *testing.T) {
	server, _ := NewHttpServer(1234, "1234", NewMockDB())

	tests := []struct {
		name string
		want string
	}{
		{
			name: "1",
			want: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/headertest", nil)
			server.stdHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).ServeHTTP(w, r)

			contentType := w.Header().Values("Content-Type")[0]
			if contentType != tt.want {
				t.Errorf("stdHeader() = %v, want %v", contentType, tt.want)
			}
		})
	}
}

func TestHttpServer_router(t *testing.T) {
	server, _ := NewHttpServer(1234, "1234", NewMockDB())

	type args struct {
		r *http.Request
	}
	type response struct {
		code int
		body string
	}
	tests := []struct {
		name string
		s    *HttpServer
		args args
		want response
	}{
		{
			name: "1",
			s:    &server,
			args: args{r: httptest.NewRequest("GET", "/invalid", nil)},
			want: response{code: http.StatusNotFound},
		},
		{
			name: "2",
			s:    &server,
			args: args{r: httptest.NewRequest("POST", "/recipe", strings.NewReader(`{"name":"BLT","ingredients":["Tomato","Bacon","Lettuce"]}`))},
			want: response{code: http.StatusOK},
		},
		{
			name: "3",
			s:    &server,
			args: args{r: httptest.NewRequest("GET", "/recipe/BLT", nil)},
			want: response{code: http.StatusOK, body: `{"name":"BLT","ingredients":["Tomato","Bacon","Lettuce"]}`},
		},
		{
			name: "4",
			s:    &server,
			args: args{r: httptest.NewRequest("GET", "/recipes?ingredients=Tomato,Bacon", nil)},
			want: response{code: http.StatusOK, body: `{"recipes":[{"name":"BLT","ingredients":["Tomato","Bacon","Lettuce"]}]}`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			tt.s.router(w, tt.args.r)

			if w.Code != tt.want.code || w.Body.String() != tt.want.body {
				t.Errorf("auth() = %v, want %v", response{code: w.Code, body: w.Body.String()}, tt.want)
			}
		})
	}
}
