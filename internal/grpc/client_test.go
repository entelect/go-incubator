package grpc

import (
	"context"
	"fmt"
	"go-incubator/internal/http"
	"go-incubator/proto"
	"log"
	"net"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

type mockServer struct{}

func (s *mockServer) AddRecipe(ctx context.Context, r *proto.Recipe) (*emptypb.Empty, error) {
	if r.Name == "expect error" {
		return &emptypb.Empty{}, status.Errorf(codes.Internal, "expected error")
	}

	return &emptypb.Empty{}, nil
}

func (s *mockServer) GetRecipe(ctx context.Context, r *proto.RecipeRequest) (*proto.Recipe, error) {
	switch r.Name {
	case "BLT":
		return &proto.Recipe{Name: "BLT", Ingredients: []string{"Bacon", "Lettuce", "Tomato"}}, nil
	case "expect error":
		return nil, status.Errorf(codes.Internal, "expected error")
	}

	return nil, status.Errorf(codes.NotFound, "recipe (%s) not found", r.Name)
}

func (s *mockServer) FindRecipes(ctx context.Context, r *proto.FindRequest) (*proto.Recipes, error) {
	switch strings.Join(r.Ingredients, " ") {
	case "expected ok":
		return &proto.Recipes{Recipes: []*proto.Recipe{{Name: "one", Ingredients: []string{"oneone", "onetwo"}}, {Name: "two", Ingredients: []string{"twoone", "twotwo"}}}}, nil
	case "expected empty":
		return &proto.Recipes{}, nil
	case "expected error":
		return nil, status.Errorf(codes.Internal, "expected error")
	}
	return &proto.Recipes{}, nil
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	proto.RegisterRecipeServiceServer(s, &mockServer{})
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func TestNewGrpcClient(t *testing.T) {
	type args struct {
		baseUrl string
		port    int
		apiKey  string
	}
	tests := []struct {
		name    string
		args    args
		want    GrpcClient
		wantErr bool
	}{
		{
			name:    "1",
			args:    args{baseUrl: "127.0.0.1", port: 1234, apiKey: "1234"},
			want:    GrpcClient{apiKey: "1234"},
			wantErr: false,
		},
		{
			name:    "2",
			args:    args{baseUrl: `<>#%"`, port: 9999999, apiKey: "1234"},
			want:    GrpcClient{apiKey: "1234"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewGrpcClient(tt.args.baseUrl, tt.args.port, tt.args.apiKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGrpcClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.apiKey, tt.want.apiKey) {
				t.Errorf("NewGrpcClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGrpcClient_auth(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := proto.NewRecipeServiceClient(conn)

	type args struct {
		ctx     context.Context
		method  string
		req     interface{}
		reply   interface{}
		cc      *grpc.ClientConn
		invoker grpc.UnaryInvoker
		opts    []grpc.CallOption
	}
	tests := []struct {
		name    string
		c       *GrpcClient
		args    args
		wantErr bool
	}{
		{
			name: "1",
			c:    &GrpcClient{client: client, apiKey: "1234"},
			args: args{
				ctx: context.Background(),
				req: proto.RecipeRequest{Name: "test"},
				invoker: func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
					md, ok := metadata.FromOutgoingContext(ctx)
					if !ok {
						t.Errorf("GrpcClient.auth() error retrieving metadata")
						return fmt.Errorf("GrpcClient.auth() error retrieving metadata")
					}

					authHeader, ok := md["x-api-key"]
					if !ok || authHeader[0] != "1234" {
						t.Errorf("authentication failed")
						return fmt.Errorf("authentication failed")
					}

					return nil
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.auth(tt.args.ctx, tt.args.method, tt.args.req, tt.args.reply, tt.args.cc, tt.args.invoker, tt.args.opts...); (err != nil) != tt.wantErr {
				t.Errorf("GrpcClient.auth() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGrpcClient_AddRecipe(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := proto.NewRecipeServiceClient(conn)

	type args struct {
		recipe http.Recipe
	}
	tests := []struct {
		name    string
		c       *GrpcClient
		args    args
		wantErr bool
	}{
		{
			name:    "1",
			c:       &GrpcClient{client: client, apiKey: "1234"},
			args:    args{recipe: http.Recipe{Name: "expect ok", Ingredients: []string{"one", "two", "three"}}},
			wantErr: false,
		},
		{
			name:    "2",
			c:       &GrpcClient{client: client, apiKey: "1234"},
			args:    args{recipe: http.Recipe{Name: "expect error", Ingredients: []string{"one", "two", "three"}}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.AddRecipe(tt.args.recipe); (err != nil) != tt.wantErr {
				t.Errorf("GrpcClient.AddRecipe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGrpcClient_GetRecipe(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := proto.NewRecipeServiceClient(conn)

	type args struct {
		name string
	}
	tests := []struct {
		name    string
		c       *GrpcClient
		args    args
		want    *http.Recipe
		wantErr bool
	}{
		{
			name:    "1",
			c:       &GrpcClient{client: client, apiKey: "1234"},
			args:    args{name: "BLT"},
			want:    &http.Recipe{Name: "BLT", Ingredients: []string{"Bacon", "Lettuce", "Tomato"}},
			wantErr: false,
		},
		{
			name:    "2",
			c:       &GrpcClient{client: client, apiKey: "1234"},
			args:    args{name: "Bobotie"},
			want:    nil,
			wantErr: false,
		},
		{
			name:    "3",
			c:       &GrpcClient{client: client, apiKey: "1234"},
			args:    args{name: "expect error"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.GetRecipe(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("GrpcClient.GetRecipe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GrpcClient.GetRecipe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGrpcClient_SearchByIngredients(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := proto.NewRecipeServiceClient(conn)

	type args struct {
		ingredients []string
	}
	tests := []struct {
		name    string
		c       *GrpcClient
		args    args
		want    []http.Recipe
		wantErr bool
	}{
		{
			name:    "1",
			c:       &GrpcClient{client: client, apiKey: "1234"},
			args:    args{ingredients: []string{"expected", "ok"}},
			want:    []http.Recipe{{Name: "one", Ingredients: []string{"oneone", "onetwo"}}, {Name: "two", Ingredients: []string{"twoone", "twotwo"}}},
			wantErr: false,
		},
		{
			name:    "2",
			c:       &GrpcClient{client: client, apiKey: "1234"},
			args:    args{ingredients: []string{"expected", "empty"}},
			want:    nil,
			wantErr: false,
		},
		{
			name:    "3",
			c:       &GrpcClient{client: client, apiKey: "1234"},
			args:    args{ingredients: []string{"expected", "error"}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.SearchByIngredients(tt.args.ingredients)
			if (err != nil) != tt.wantErr {
				t.Errorf("GrpcClient.SearchByIngredients() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GrpcClient.SearchByIngredients() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGrpcClient_Benchmarks(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := proto.NewRecipeServiceClient(conn)

	type args struct {
		duration time.Duration
	}
	tests := []struct {
		name      string
		c         *GrpcClient
		args      args
		wantRegex string
	}{
		{
			name: "1",
			c:    &GrpcClient{client: client, apiKey: "1234"},
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
				t.Errorf("GrpcClient.Benchmarks(): %v", err)
			}
			if !match {
				t.Errorf("GrpcClient.Benchmarks() = %v, want %v", got, tt.wantRegex)
			}
		})
	}
}
