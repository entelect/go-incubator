package http

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

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
				apiKey:  "1234",
				recipes: make(map[string]Recipe),
			},
		},
		{
			name: "2",
			args: args{port: 1234, apiKey: ""},
			want: HttpServer{
				server: &http.Server{
					Addr: fmt.Sprintf(":%d", 1234),
				},
				apiKey:  "",
				recipes: make(map[string]Recipe),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewHttpServer(tt.args.port, tt.args.apiKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHttpServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.server.Addr != tt.want.server.Addr || !reflect.DeepEqual(got.recipes, tt.want.recipes) {
				t.Errorf("NewHttpServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHttpServer_addRecipe(t *testing.T) {
	server, _ := NewHttpServer(1234, "1234")

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
	server, _ := NewHttpServer(1234, "1234")
	server.recipes["Cheese Fondue"] = Recipe{Name: "Cheese Fondue", Ingredients: []string{"Gruyere", "Emmental"}}
	server.recipes["Mac & Cheese"] = Recipe{Name: "Mac & Cheese", Ingredients: []string{"Mozzarella", "Macaroni"}}
	server.recipes["SpagBol"] = Recipe{Name: "SpagBol", Ingredients: []string{"Spaghetti", "Ground Beef", "Tomato"}}
	server.recipes["BLT"] = Recipe{Name: "BLT", Ingredients: []string{"Tomato", "Bacon", "Lettuce"}}
	server.recipes["Greek Salad"] = Recipe{Name: "Greek Salad", Ingredients: []string{"Feta", "Tomato", "Cucumber"}}
	server.recipes["Caprese Salad"] = Recipe{Name: "Caprese Salad", Ingredients: []string{"Mozzarella", "Tomato"}}
	server.recipes["Meatballs"] = Recipe{Name: "Meatballs", Ingredients: []string{"Ground Beef", "Tomato"}}

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
	server, _ := NewHttpServer(1234, "1234")
	server.recipes["Cheese Fondue"] = Recipe{Name: "Cheese Fondue", Ingredients: []string{"Gruyere", "Emmental"}}
	server.recipes["Mac & Cheese"] = Recipe{Name: "Mac & Cheese", Ingredients: []string{"Mozzarella", "Macaroni"}}
	server.recipes["SpagBol"] = Recipe{Name: "SpagBol", Ingredients: []string{"Spaghetti", "Ground Beef", "Tomato"}}
	server.recipes["BLT"] = Recipe{Name: "BLT", Ingredients: []string{"Tomato", "Bacon", "Lettuce"}}
	server.recipes["Greek Salad"] = Recipe{Name: "Greek Salad", Ingredients: []string{"Feta", "Tomato", "Cucumber"}}
	server.recipes["Caprese Salad"] = Recipe{Name: "Caprese Salad", Ingredients: []string{"Mozzarella", "Tomato"}}
	server.recipes["Meatballs"] = Recipe{Name: "Meatballs", Ingredients: []string{"Ground Beef", "Tomato"}}

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
				body: `[{"name":"Cheese Fondue","ingredients":["Gruyere","Emmental"]}]`,
			},
		},
		{
			name: "2",
			path: "/recipes?ingredients=Emmental,Gruyere",
			want: response{
				code: http.StatusOK,
				body: `[{"name":"Cheese Fondue","ingredients":["Gruyere","Emmental"]}]`,
			},
		},
		{
			name: "3",
			path: "/recipes?ingredients=Tomato",
			want: response{
				code: http.StatusOK,
				body: `[{"name":"BLT","ingredients":["Tomato","Bacon","Lettuce"]},{"name":"Caprese Salad","ingredients":["Mozzarella","Tomato"]},{"name":"Greek Salad","ingredients":["Feta","Tomato","Cucumber"]},{"name":"Meatballs","ingredients":["Ground Beef","Tomato"]},{"name":"SpagBol","ingredients":["Spaghetti","Ground Beef","Tomato"]}]`,
			},
		},
		{
			name: "4",
			path: "/recipes?ingredients=Tomato,Onion",
			want: response{
				code: http.StatusOK,
				body: `[]`,
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
	server, _ := NewHttpServer(1234, "1234")

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
	server, _ := NewHttpServer(1234, "1234")

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
