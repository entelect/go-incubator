package grpc

import (
	"context"
	"fmt"
	"go-incubator/internal/persistence"
	"go-incubator/proto"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GrpcServer struct {
	server *grpc.Server
	port   int
	apiKey string
	db     persistence.Persistence
}

// NewGrpcServer creates and returns a new GrpcServer with a listener on the specified port
func NewGrpcServer(port int, apiKey string, persistence persistence.Persistence) (GrpcServer, error) {
	s := GrpcServer{
		port:   port,
		apiKey: apiKey,
		db:     persistence,
	}

	return s, nil
}

// Start initiates the gRPC listener of the received GrpcServer
func (s *GrpcServer) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
		if err != nil {
			fmt.Printf("tcp listener error: %v\n", err)
			return
		}

		s.server = grpc.NewServer()
		proto.RegisterRecipeServiceServer(s.server, &serviceServer{db: s.db})

		fmt.Printf("starting gRPC listener on port %d\n", s.port)
		defer fmt.Printf("gRPC listener on port %d stopped\n", s.port)
		if err := s.server.Serve(lis); err != nil {
			fmt.Printf("gRPC server error: %v\n", err)
			return
		}
	}()
}

// Stop terminates the gRPC listener of the received GrpcServer
func (s *GrpcServer) Stop() {
	if s.server != nil {
		s.server.Stop()
	}
}

// server is used to implement RecipeServiceServer
type serviceServer struct {
	db persistence.Persistence
}

func (s *serviceServer) AddRecipe(ctx context.Context, r *proto.Recipe) (*emptypb.Empty, error) {
	if r.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "no name specified")
	}

	if len(r.Ingredients) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "no ingredients specified")
	}

	// Convert *proto.Recipe to persistence.Recipe
	recipe := persistence.Recipe{}
	recipe.Name = r.Name
	recipe.Ingredients = r.Ingredients

	err := s.db.AddRecipe(recipe)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "writing recipe to db: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *serviceServer) GetRecipe(ctx context.Context, r *proto.RecipeRequest) (*proto.Recipe, error) {
	recipe, err := s.db.GetRecipe(r.Name)
	if err == persistence.ErrNoResults {
		return nil, status.Errorf(codes.NotFound, "recipe (%s) not found", r.Name)
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "getting recipe from db: %v", err)
	}

	// Convert persistence.Recipe to *proto.Recipe
	rsp := &proto.Recipe{}
	rsp.Name = recipe.Name
	rsp.Ingredients = recipe.Ingredients

	return rsp, nil
}

func (s *serviceServer) FindRecipes(ctx context.Context, r *proto.FindRequest) (*proto.Recipes, error) {
	if len(r.Ingredients) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "no ingredients specified")
	}

	dbrecipes, err := s.db.FindRecipes(r.Ingredients)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "reading recipes from db: %v", err)
	}

	// Convert []persistence.Recipe to *proto.Recipes
	rsp := &proto.Recipes{Recipes: []*proto.Recipe{}}
	for _, r := range dbrecipes {
		recipe := &proto.Recipe{}
		recipe.Name = r.Name
		recipe.Ingredients = r.Ingredients
		rsp.Recipes = append(rsp.Recipes, recipe)
	}

	return rsp, nil
}
