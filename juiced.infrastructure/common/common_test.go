package common

import (
	"testing"
)

func TestRandString(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		args args
		want int
	}{
		{args: args{n: 16}, want: 16},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := len(RandString(tt.args.n)); got != tt.want {
				t.Errorf("RandString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRandID(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		args args
		want int
	}{
		{args: args{n: 16}, want: 16},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := len(RandID(tt.args.n)); got != tt.want {
				t.Errorf("RandID() = %v, want %v", got, tt.want)
			}
		})
	}
}
