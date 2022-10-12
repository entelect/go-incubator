package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestNewHttpServer(t *testing.T) {
	type args struct {
		port int
	}
	tests := []struct {
		name    string
		args    args
		want    HttpServer
		wantErr bool
	}{
		{
			name: "1",
			args: args{port: 1234},
			want: HttpServer{
				server: &http.Server{
					Addr: fmt.Sprintf(":%d", 1234),
				},
				recipes: make(map[string]Recipe),
			},
		},
		{
			name: "2",
			args: args{port: 1234},
			want: HttpServer{
				server: &http.Server{
					Addr: fmt.Sprintf(":%d", 1234),
				},
				recipes: make(map[string]Recipe),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewHttpServer(tt.args.port)
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
	server, _ := NewHttpServer(1234)

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
	server, _ := NewHttpServer(1234)
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
	server, _ := NewHttpServer(1234)
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
