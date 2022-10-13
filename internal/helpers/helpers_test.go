package helpers

import (
	"testing"
)

func Test_stringSlicesEqual(t *testing.T) {
	type args struct {
		a          []string
		b          []string
		checkOrder bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "1",
			args: args{
				a:          []string{"one", "two", "three"},
				b:          []string{"one", "two", "three"},
				checkOrder: true,
			},
			want: true,
		},
		{
			name: "2",
			args: args{
				a:          []string{"one", "two", "three"},
				b:          []string{"one", "two", "three"},
				checkOrder: false,
			},
			want: true,
		},
		{
			name: "3",
			args: args{
				a:          []string{"one", "two", "three"},
				b:          []string{"three", "two", "one"},
				checkOrder: true,
			},
			want: false,
		},
		{
			name: "4",
			args: args{
				a:          []string{"one", "two", "three"},
				b:          []string{"three", "two", "one"},
				checkOrder: false,
			},
			want: true,
		},
		{
			name: "5",
			args: args{
				a:          []string{"one", "two", "three"},
				b:          []string{"one", "two", "three", "four"},
				checkOrder: true,
			},
			want: false,
		},
		{
			name: "6",
			args: args{
				a:          []string{"one", "two", "three"},
				b:          []string{"one", "two", "three", "four"},
				checkOrder: false,
			},
			want: false,
		},
		{
			name: "7",
			args: args{
				a:          []string{"one", "two", "three", "four"},
				b:          []string{"one", "two", "three"},
				checkOrder: true,
			},
			want: false,
		},
		{
			name: "8",
			args: args{
				a:          []string{},
				b:          []string{},
				checkOrder: true,
			},
			want: true,
		},
		{
			name: "8",
			args: args{
				a:          []string{},
				b:          []string{},
				checkOrder: false,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringSlicesEqual(tt.args.a, tt.args.b, tt.args.checkOrder); got != tt.want {
				t.Errorf("stringSlicesEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_stringSliceContains(t *testing.T) {
	type args struct {
		a []string
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "1",
			args: args{
				a: []string{"one", "two", "three"},
				s: "one",
			},
			want: true,
		},
		{
			name: "2",
			args: args{
				a: []string{"one", "two", "three"},
				s: "four",
			},
			want: false,
		},
		{
			name: "3",
			args: args{
				a: []string{"one", "two", "three"},
				s: "ONE",
			},
			want: false,
		},
		{
			name: "4",
			args: args{
				a: []string{},
				s: "one",
			},
			want: false,
		},
		{
			name: "5",
			args: args{
				a: []string{"one", "two", "three"},
				s: "",
			},
			want: false,
		},
		{
			name: "6",
			args: args{
				a: []string{},
				s: "",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringSliceContains(tt.args.a, tt.args.s); got != tt.want {
				t.Errorf("stringSliceContains() = %v, want %v", got, tt.want)
			}
		})
	}
}
