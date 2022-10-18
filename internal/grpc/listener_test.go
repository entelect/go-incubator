package grpc

import (
	"bytes"
	"context"
	"fmt"
	"go-incubator/internal/persistence"
	"go-incubator/proto"
	"io"
	"log"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	pb "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
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
	if recipe.Name == "Expected Error" {
		return fmt.Errorf("database error")
	}
	return nil
}

func (db *mockdb) GetRecipe(name string) (persistence.Recipe, error) {
	if name == "Expected Error" {
		return persistence.Recipe{}, fmt.Errorf("database error")
	}
	r, ok := db.recipes[name]
	if !ok {
		return persistence.Recipe{}, persistence.ErrNoResults
	}

	return r, nil
}

func (db *mockdb) FindRecipes(ingredients []string) ([]persistence.Recipe, error) {
	if strings.Join(ingredients, " ") == "Expected Error" {
		return nil, fmt.Errorf("database error")
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

func TestNewGrpcServer(t *testing.T) {
	type args struct {
		port        int
		apiKey      string
		persistence persistence.Persistence
	}
	tests := []struct {
		name    string
		args    args
		want    GrpcServer
		wantErr bool
	}{
		{
			name: "1",
			args: args{
				port:        1234,
				apiKey:      "1234",
				persistence: NewMockDB(),
			},
			want: GrpcServer{port: 1234, apiKey: "1234", db: NewMockDB()},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewGrpcServer(tt.args.port, tt.args.apiKey, tt.args.persistence)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGrpcServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewGrpcServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGrpcServer_auth(t *testing.T) {
	type args struct {
		ctx     context.Context
		req     interface{}
		info    *grpc.UnaryServerInfo
		handler grpc.UnaryHandler
	}
	tests := []struct {
		name    string
		s       *GrpcServer
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "1",
			s:    &GrpcServer{apiKey: "1234", db: NewMockDB()},
			args: args{
				ctx:     metadata.NewIncomingContext(context.Background(), metadata.MD{"x-api-key": []string{"1234"}}),
				req:     "request1",
				handler: func(ctx context.Context, req interface{}) (interface{}, error) { return nil, nil },
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "2",
			s:    &GrpcServer{apiKey: "1111", db: NewMockDB()},
			args: args{
				ctx:     metadata.NewIncomingContext(context.Background(), metadata.MD{"x-api-key": []string{"1234"}}),
				req:     "request2",
				handler: func(ctx context.Context, req interface{}) (interface{}, error) { return nil, nil },
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "3",
			s:    &GrpcServer{apiKey: "1234", db: NewMockDB()},
			args: args{
				ctx:     context.Background(),
				req:     "request3",
				handler: func(ctx context.Context, req interface{}) (interface{}, error) { return nil, nil },
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.auth(tt.args.ctx, tt.args.req, tt.args.info, tt.args.handler)
			if (err != nil) != tt.wantErr {
				t.Errorf("GrpcServer.auth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GrpcServer.auth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGrpcServer_tracer(t *testing.T) {
	type args struct {
		ctx     context.Context
		req     interface{}
		info    *grpc.UnaryServerInfo
		handler grpc.UnaryHandler
	}
	tests := []struct {
		name      string
		s         *GrpcServer
		args      args
		wantRegex string
	}{
		{
			name: "1",
			s:    &GrpcServer{apiKey: "1234", db: NewMockDB()},
			args: args{ctx: context.Background(), req: "request1", info: &grpc.UnaryServerInfo{FullMethod: "test method"}, handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			}},
			wantRegex: `test method 0(.\d+)?s\n`,
		},
		{
			name: "2",
			s:    &GrpcServer{apiKey: "1234", db: NewMockDB()},
			args: args{ctx: context.Background(), req: "request1", info: &grpc.UnaryServerInfo{FullMethod: "test method"}, handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				time.Sleep(5 * time.Second)
				return nil, nil
			}},
			wantRegex: `test method 5(.\d+)?s\n`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := captureOutput(func() { tt.s.tracer(tt.args.ctx, tt.args.req, tt.args.info, tt.args.handler) })

			match, err := regexp.MatchString(tt.wantRegex, got)
			if err != nil {
				t.Errorf("GrpcServer.tracer(): %v", err)
			}
			if !match {
				t.Errorf("GrpcServer.tracer() = %v, want %v", got, tt.wantRegex)
			}
		})
	}
}

func Test_serviceServer_AddRecipe(t *testing.T) {
	type args struct {
		ctx context.Context
		r   *proto.Recipe
	}
	tests := []struct {
		name    string
		s       *serviceServer
		args    args
		want    *emptypb.Empty
		wantErr bool
	}{
		{
			name: "1",
			s:    &serviceServer{db: NewMockDB()},
			args: args{
				ctx: context.Background(),
				r:   &proto.Recipe{Name: "Meatballs", Ingredients: []string{"Ground Beef", "Tomato"}},
			},
			want:    &emptypb.Empty{},
			wantErr: false,
		},
		{
			name: "2",
			s:    &serviceServer{db: NewMockDB()},
			args: args{
				ctx: context.Background(),
				r:   &proto.Recipe{Ingredients: []string{"Mozzarella", "Macaroni"}},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "3",
			s:    &serviceServer{db: NewMockDB()},
			args: args{
				ctx: context.Background(),
				r:   &proto.Recipe{Name: "Pizza"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "4",
			s:    &serviceServer{db: NewMockDB()},
			args: args{
				ctx: context.Background(),
				r:   &proto.Recipe{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "5",
			s:    &serviceServer{db: NewMockDB()},
			args: args{
				ctx: context.Background(),
				r:   &proto.Recipe{Name: "Expected Error", Ingredients: []string{"Expected", "Error"}},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.AddRecipe(tt.args.ctx, tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("serviceServer.AddRecipe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !pb.Equal(got, tt.want) {
				t.Errorf("serviceServer.AddRecipe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_serviceServer_GetRecipe(t *testing.T) {
	type args struct {
		ctx context.Context
		r   *proto.RecipeRequest
	}
	tests := []struct {
		name    string
		s       *serviceServer
		args    args
		want    *proto.Recipe
		wantErr bool
	}{
		{
			name:    "1",
			s:       &serviceServer{db: NewMockDB()},
			args:    args{ctx: context.Background(), r: &proto.RecipeRequest{Name: "Cheese Fondue"}},
			want:    &proto.Recipe{Name: "Cheese Fondue", Ingredients: []string{"Gruyere", "Emmental"}},
			wantErr: false,
		},
		{
			name:    "2",
			s:       &serviceServer{db: NewMockDB()},
			args:    args{ctx: context.Background(), r: &proto.RecipeRequest{Name: "SpagBol"}},
			want:    &proto.Recipe{Name: "SpagBol", Ingredients: []string{"Spaghetti", "Ground Beef", "Tomato"}},
			wantErr: false,
		},
		{
			name:    "3",
			s:       &serviceServer{db: NewMockDB()},
			args:    args{ctx: context.Background(), r: &proto.RecipeRequest{Name: "Pizza"}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "4",
			s:       &serviceServer{db: NewMockDB()},
			args:    args{ctx: context.Background(), r: &proto.RecipeRequest{Name: "Expected Error"}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.GetRecipe(tt.args.ctx, tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("serviceServer.GetRecipe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("serviceServer.GetRecipe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_serviceServer_FindRecipes(t *testing.T) {
	type args struct {
		ctx context.Context
		r   *proto.FindRequest
	}
	tests := []struct {
		name    string
		s       *serviceServer
		args    args
		want    *proto.Recipes
		wantErr bool
	}{
		{
			name:    "1",
			s:       &serviceServer{db: NewMockDB()},
			args:    args{ctx: context.Background(), r: &proto.FindRequest{Ingredients: []string{"Gruyere", "Emmental"}}},
			want:    &proto.Recipes{Recipes: []*proto.Recipe{{Name: "Cheese Fondue", Ingredients: []string{"Gruyere", "Emmental"}}}},
			wantErr: false,
		},
		{
			name:    "2",
			s:       &serviceServer{db: NewMockDB()},
			args:    args{ctx: context.Background(), r: &proto.FindRequest{Ingredients: []string{"Emmental", "Gruyere"}}},
			want:    &proto.Recipes{Recipes: []*proto.Recipe{{Name: "Cheese Fondue", Ingredients: []string{"Gruyere", "Emmental"}}}},
			wantErr: false,
		},
		{
			name:    "3",
			s:       &serviceServer{db: NewMockDB()},
			args:    args{ctx: context.Background(), r: &proto.FindRequest{Ingredients: []string{"Tomato"}}},
			want:    &proto.Recipes{Recipes: []*proto.Recipe{{Name: "BLT", Ingredients: []string{"Tomato", "Bacon", "Lettuce"}}, {Name: "Caprese Salad", Ingredients: []string{"Mozzarella", "Tomato"}}, {Name: "Greek Salad", Ingredients: []string{"Feta", "Tomato", "Cucumber"}}, {Name: "Meatballs", Ingredients: []string{"Ground Beef", "Tomato"}}, {Name: "SpagBol", Ingredients: []string{"Spaghetti", "Ground Beef", "Tomato"}}}},
			wantErr: false,
		},
		{
			name:    "4",
			s:       &serviceServer{db: NewMockDB()},
			args:    args{ctx: context.Background(), r: &proto.FindRequest{Ingredients: []string{"Tomato", "Onion"}}},
			want:    &proto.Recipes{Recipes: []*proto.Recipe{}},
			wantErr: false,
		},
		{
			name:    "5",
			s:       &serviceServer{db: NewMockDB()},
			args:    args{ctx: context.Background(), r: &proto.FindRequest{}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "6",
			s:       &serviceServer{db: NewMockDB()},
			args:    args{ctx: context.Background(), r: &proto.FindRequest{Ingredients: []string{"Expected", "Error"}}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.FindRecipes(tt.args.ctx, tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("serviceServer.FindRecipes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("serviceServer.FindRecipes() = %v, want %v", got, tt.want)
			}
		})
	}
}
