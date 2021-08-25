package common

import (
	"reflect"
	"strings"
	"testing"

	"backend.juicedbot.io/juiced.infrastructure/common/enums"
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

func TestCreateParams(t *testing.T) {
	type args struct {
		paramsLong map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "One", args: args{paramsLong: map[string]string{"ONE": "TRUE"}}, want: "ONE=TRUE"},
		{name: "Two", args: args{paramsLong: map[string]string{"ONE": "TRUE", "TWO": "TRUE"}}, want: "ONE=TRUE&TWO=TRUE"},
		{name: "Three", args: args{paramsLong: map[string]string{"ONE": "TRUE", "TWO": "TRUE", "THREE": "TRUE"}}, want: "ONE=TRUE&TWO=TRUE&THREE=TRUE"},
		{name: "Three Bad", args: args{paramsLong: map[string]string{"ONE": "TRUE", "TWO": "TRUE", "THREE": "TRUE"}}, want: "ONE=TRUE&TWO=TRUE&THREE=WRONG", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			success := true
			params := make(map[string]string)
			badParams := []string{}
			got := CreateParams(tt.args.paramsLong)
			splitted1 := strings.Split(got, "&")
			for _, split1 := range splitted1 {
				splitted2 := strings.Split(split1, "=")
				params[splitted2[0]] = splitted2[1]
			}
			wantSplitted1 := strings.Split(tt.want, "&")
			for _, wantSplit1 := range wantSplitted1 {
				wantSplitted2 := strings.Split(wantSplit1, "=")
				param, ok := params[wantSplitted2[0]]
				if !ok {
					success = false
					badParams = append(badParams, wantSplitted2[0])
					break
				}
				if param != wantSplitted2[1] {
					success = false
					badParams = append(badParams, wantSplitted2[0])
					break
				}
			}
			if !success && !tt.wantErr {
				t.Errorf("CreateParams() returned wrong key(s) %v", badParams)
			}
		})
	}
}

func TestValidCardType(t *testing.T) {
	type args struct {
		cardNumber []byte
		retailer   enums.Retailer
	}
	tests := []struct {
		args args
		want bool
	}{
		{
			args: args{
				cardNumber: []byte("5859254368973596"),
				retailer:   enums.Target,
			},
			want: true,
		},
		{
			args: args{
				cardNumber: []byte("6394254368973596"),
				retailer:   enums.Target,
			},
			want: true,
		},
		{
			args: args{
				cardNumber: []byte("6395254368973596"),
				retailer:   enums.Target,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := ValidCardType(tt.args.cardNumber, tt.args.retailer); got != tt.want {
				t.Errorf("ValidCardType() = %v, want %v", got, tt.want)
			}
		})
	}
}
