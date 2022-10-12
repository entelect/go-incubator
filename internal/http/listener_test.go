package http

import (
	"fmt"
	"net/http"
	"reflect"
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
