package common

import (
	"reflect"
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

func TestRemoveFromSlice(t *testing.T) {
	type args struct {
		s []string
		x string
	}
	tests := []struct {
		args args
		want []string
	}{
		{args: args{s: []string{"foo", "bar", "far"}, x: "foo"}, want: []string{"bar", "far"}},
		{args: args{s: []string{"foo", "bar", "far"}, x: "bar"}, want: []string{"foo", "far"}},
		{args: args{s: []string{"foo", "bar", "far"}, x: "far"}, want: []string{"foo", "bar"}},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := RemoveFromSlice(tt.args.s, tt.args.x)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveFromSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
