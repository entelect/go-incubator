package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestNewHttpClient(t *testing.T) {
	type args struct {
		baseUrl string
		port    int
		apiKey  string
	}
	tests := []struct {
		name    string
		args    args
		want    HttpClient
		wantErr bool
	}{
		{
			name: "1",
			args: args{baseUrl: "127.0.0.1", port: 1234, apiKey: "1234"},
			want: HttpClient{
				client:  &http.Client{},
				address: "http://127.0.0.1:1234",
				apiKey:  "1234",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewHttpClient(tt.args.baseUrl, tt.args.port, tt.args.apiKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHttpClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHttpClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHttpClient_AddRecipe(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(""))
	}))
	defer server.Close()

	client := HttpClient{
		client:  &http.Client{},
		address: server.URL,
		apiKey:  "1234",
	}

	tests := []struct {
		name    string
		c       *HttpClient
		recipe  Recipe
		wantErr bool
	}{
		{
			name:    "1",
			c:       &client,
			recipe:  Recipe{Name: "one", Ingredients: []string{"one", "two", "three"}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.AddRecipe(tt.recipe); (err != nil) != tt.wantErr {
				t.Errorf("HttpClient.AddRecipe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHttpClient_GetRecipe(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name, err := url.QueryUnescape(strings.TrimPrefix(r.RequestURI, "/recipe/"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		switch name {
		case "notfound":
			w.WriteHeader(http.StatusNotFound)
		case "found":
			rsp, _ := json.Marshal(Recipe{
				Name:        "found",
				Ingredients: []string{"one", "two", "three"},
			})
			w.WriteHeader(http.StatusOK)
			w.Write(rsp)
		case "badgateway":
			w.WriteHeader(http.StatusBadGateway)
		}
	}))
	defer server.Close()

	client := HttpClient{
		client:  &http.Client{},
		address: server.URL,
		apiKey:  "1234",
	}

	tests := []struct {
		name    string
		c       *HttpClient
		rname   string
		want    *Recipe
		wantErr error
	}{
		{
			name:    "1",
			c:       &client,
			rname:   "notfound",
			want:    nil,
			wantErr: nil,
		},
		{
			name:    "2",
			c:       &client,
			rname:   "badgateway",
			want:    nil,
			wantErr: fmt.Errorf("502 Bad Gateway"),
		},
		{
			name:    "3",
			c:       &client,
			rname:   "found",
			want:    &Recipe{Name: "found", Ingredients: []string{"one", "two", "three"}},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.GetRecipe(tt.rname)
			if (err == nil) != (tt.wantErr == nil) {
				t.Errorf("HttpClient.GetRecipe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr != nil && (err.Error() != tt.wantErr.Error()) {
				t.Errorf("HttpClient.GetRecipe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HttpClient.GetRecipe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHttpClient_SearchByIngredients(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		unescaped, err := url.QueryUnescape(strings.TrimPrefix(r.RequestURI, "/recipes"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		elems := strings.Split(unescaped, "?")
		if len(elems) > 2 {
			w.WriteHeader((http.StatusBadRequest))
		}

		var params []string
		if len(elems) > 1 {
			params = strings.Split(elems[1], "&")
		}

		var ingredients string
		for _, v := range params {
			if strings.HasPrefix(v, "ingredients=") {
				ingredients = strings.TrimPrefix(v, "ingredients=")
			}
		}

		switch ingredients {
		case "Tomato,Ground Beef":
			recipes := `{"recipes":[{"name":"Meatballs","ingredients":["Ground Beef","Tomato"]},{"name":"SpagBol","ingredients":["Spaghetti","Ground Beef","Tomato"]}]}`
			w.Write([]byte(recipes))
			w.WriteHeader(http.StatusOK)
		case "Tomato,Ground Beef,Spaghetti":
			recipes := `{"recipes":[{"name":"SpagBol","ingredients":["Spaghetti","Ground Beef","Tomato"]}]}`
			w.Write([]byte(recipes))
			w.WriteHeader(http.StatusOK)
		case "Tomato,Ground Beef,Spaghetti,Crickets":
			recipes := `{"recipes":[]}`
			w.Write([]byte(recipes))
			w.WriteHeader(http.StatusOK)
		case "":
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	client := HttpClient{
		client:  &http.Client{},
		address: server.URL,
		apiKey:  "1234",
	}

	tests := []struct {
		name        string
		c           *HttpClient
		ingredients []string
		want        []Recipe
		wantErr     error
	}{
		{
			name:        "1",
			c:           &client,
			ingredients: []string{"Tomato", "Ground Beef"},
			want: []Recipe{
				{Name: "Meatballs", Ingredients: []string{"Ground Beef", "Tomato"}},
				{Name: "SpagBol", Ingredients: []string{"Spaghetti", "Ground Beef", "Tomato"}},
			},
			wantErr: nil,
		},
		{
			name:        "2",
			c:           &client,
			ingredients: []string{"Tomato", "Ground Beef", "Spaghetti"},
			want: []Recipe{
				{Name: "SpagBol", Ingredients: []string{"Spaghetti", "Ground Beef", "Tomato"}},
			},
			wantErr: nil,
		},
		{
			name:        "3",
			c:           &client,
			ingredients: []string{"Tomato", "Ground Beef", "Spaghetti", "Crickets"},
			want:        []Recipe{},
			wantErr:     nil,
		},
		{
			name:        "4",
			c:           &client,
			ingredients: []string{},
			want:        nil,
			wantErr:     fmt.Errorf("400 Bad Request"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.SearchByIngredients(tt.ingredients)
			if (err == nil) != (tt.wantErr == nil) {
				t.Errorf("HttpClient.SearchByIngredients() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr != nil && (err.Error() != tt.wantErr.Error()) {
				t.Errorf("HttpClient.SearchByIngredients() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HttpClient.SearchByIngredients() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHttpClient_Benchmarks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()

	client := HttpClient{
		client:  &http.Client{},
		address: server.URL,
		apiKey:  "1234",
	}

	type args struct {
		duration time.Duration
	}
	tests := []struct {
		name      string
		c         *HttpClient
		args      args
		wantRegex string
	}{
		{
			name: "1",
			c:    &client,
			args: args{duration: 1 * time.Second},
			wantRegex: `Calling SearchByIngredients\(\[\]string\{"Tomato"\}\) on 100 concurrent routines for 1s, please wait
┎──────────────────────────────────────────────────────────┒
\.+
SearchByIngredients\(\[\]string\{"Tomato"\}\) called \d+ times in 1s`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := captureOutput(func() { tt.c.Benchmarks(tt.args.duration) })

			match, err := regexp.MatchString(tt.wantRegex, got)
			if err != nil {
				t.Errorf("HttpClient.Benchmarks(): %v", err)
			}
			if !match {
				t.Errorf("HttpClient.Benchmarks() = %v, want %v", got, tt.wantRegex)
			}
		})
	}
}
