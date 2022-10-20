package hybrid

import (
	"context"
	"fmt"
	"go-incubator/internal/persistence"
	"go-incubator/proto"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type HybridServer struct {
	grpcServer *grpc.Server
	httpServer *http.Server
	httpPort   int
	grpcPort   int
	apiKey     string
	db         persistence.Persistence
}

// NewHybridServer creates and returns a new HybridServer with a listener on the specified port
func NewHybridServer(httpPort int, grpcPort int, apiKey string, persistence persistence.Persistence) (HybridServer, error) {
	s := HybridServer{
		httpPort: httpPort,
		grpcPort: grpcPort,
		apiKey:   apiKey,
		db:       persistence,
	}

	return s, nil
}

// Start initiates the gRPC and HTTP listeners of the received HybridServer
func (s *HybridServer) Start(wg *sync.WaitGroup) {
	// Spin up Goroutine for GRPC server
	wg.Add(1)
	go func() {
		defer wg.Done()

		tcpListener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.grpcPort))
		if err != nil {
			fmt.Printf("tcp listener error: %v\n", err)
		}

		// Set up grpc server
		s.grpcServer = grpc.NewServer(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(s.tracer, s.auth)))
		proto.RegisterRecipeServiceServer(s.grpcServer, &serviceServer{db: s.db})

		fmt.Printf("starting gRPC listener on port %d\n", s.grpcPort)
		defer fmt.Printf("gRPC listener on port %d stopped\n", s.grpcPort)
		if err := s.grpcServer.Serve(tcpListener); err != nil {
			fmt.Printf("gRPC server error: %v\n", err)
		}
	}()

	// Spin up Goroutine for REST server
	wg.Add(1)
	go func() {
		defer wg.Done()

		s.httpServer = &http.Server{
			Addr: fmt.Sprintf(":%d", s.httpPort),
		}

		mux := runtime.NewServeMux(runtime.WithIncomingHeaderMatcher(func(s string) (string, bool) {
			if strings.ToLower(s) == "x-api-key" {
				return s, true
			}
			return runtime.DefaultHeaderMatcher(s)
		}))

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		err := proto.RegisterRecipeServiceHandlerFromEndpoint(
			ctx, mux,
			fmt.Sprintf("127.0.0.1:%d", s.grpcPort),
			[]grpc.DialOption{grpc.WithInsecure()},
		)
		if err != nil {
			fmt.Printf("error layering REST server onto gRPC server: %v\n", err)
			return
		}

		s.httpServer.Handler = mux

		fmt.Printf("starting HTTP listener on port %d\n", s.httpPort)
		defer fmt.Printf("HTTP listener on port %d stopped\n", s.httpPort)
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Printf("http server error: %v\n", err)
		}
	}()
}

// Stop terminates the gRPC listener of the received GrpcServer
func (s *HybridServer) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}

	ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := s.httpServer.Shutdown(ctxTimeout); err != nil {
		panic(err) // failure/timeout shutting down the HTTP server gracefully
	}
}

// auth checks that API requests contain required API key
func (s *HybridServer) auth(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Internal, "error retrieving metadata")
	}

	authHeader, ok := md["x-api-key"]
	if !ok || authHeader[0] != s.apiKey {
		return nil, status.Error(codes.Unauthenticated, "authentication failed")
	}

	return handler(ctx, req)
}

// tracer measures the time it took for each API call to be processed
func (s *HybridServer) tracer(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	defer func(start time.Time) {
		fmt.Println(info.FullMethod, time.Since(start))
	}(time.Now())

	return handler(ctx, req)
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

	if len(r.Ingredients) == 1 {
		r.Ingredients = strings.Split(r.Ingredients[0], ",")
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
