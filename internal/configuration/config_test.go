package config

import (
	"os"
	"reflect"
	"testing"
)

func TestReadConfig(t *testing.T) {
	os.Setenv("TEST_ADDRESS", "1.1.1.1")
	os.Setenv("TEST_HTTPPORT", "1234")
	os.Setenv("TEST_GRPCPORT", "4321")
	os.Setenv("TEST_APIKEY", "1234")
	os.Setenv("TEST_DBMS", "inmem")
	os.Setenv("INVALID_HTTPPORT", "abcd")

	type args struct {
		prefix string
	}
	tests := []struct {
		name    string
		args    args
		want    Configuration
		wantErr bool
	}{
		{
			name:    "1",
			args:    args{prefix: "MISSING_"},
			want:    Configuration{},
			wantErr: false,
		},
		{
			name:    "2",
			args:    args{"INVALID_"},
			want:    Configuration{},
			wantErr: true,
		},
		{
			name: "3",
			args: args{"TEST_"},
			want: Configuration{
				HttpPort: 1234,
				GrpcPort: 4321,
				ApiKey:   "1234",
				Database: DBConfig{DBMS: "inmem"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadConfig(tt.args.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
