package logging

import (
	"reflect"
	"testing"

	"backend.juicedbot.io/juiced.client/http"
)

func TestCompareHeaders(t *testing.T) {
	type args struct {
		h1 http.Header
		h2 http.Header
	}
	tests := []struct {
		name string
		args args
		want http.Header
	}{
		{name: "Same", args: args{h1: http.Header{"header1": {"value1"}, "header2": {"value2"}}, h2: http.Header{"header1": {"value1"}, "header2": {"value2"}}}, want: http.Header{}},
		{name: "Different", args: args{h1: http.Header{"header1": {"value1"}, "header2": {"value2"}}, h2: http.Header{"header1": {"value1"}, "header2": {"different"}}}, want: http.Header{"header2": {"different"}}},
		{name: "Shorter1", args: args{h1: http.Header{"header1": {"value1"}}, h2: http.Header{"header1": {"value1"}, "header2": {"value2"}}}, want: http.Header{"header2": {"value2"}}},
		{name: "Shorter2", args: args{h1: http.Header{"header1": {"value1"}, "header2": {"value2"}}, h2: http.Header{"header1": {"value1"}}}, want: http.Header{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compareHeaders(tt.args.h1, tt.args.h2); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("compareHeaders() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareRawHeaders(t *testing.T) {
	type args struct {
		h1 http.RawHeader
		h2 http.RawHeader
	}
	tests := []struct {
		name string
		args args
		want http.RawHeader
	}{
		{"Same", args{h1: http.RawHeader{{"header1", "value1"}, {"header2", "value2"}}, h2: http.RawHeader{{"header1", "value1"}, {"header2", "value2"}}}, http.RawHeader{}},
		{"Different", args{h1: http.RawHeader{{"header1", "value1"}, {"header2", "value2"}}, h2: http.RawHeader{{"header1", "different"}, {"header2", "value2"}}}, http.RawHeader{{"header1", "different"}}},
		{"Shorter1", args{h1: http.RawHeader{{"header1", "value1"}}, h2: http.RawHeader{{"header1", "value1"}, {"header2", "value2"}}}, http.RawHeader{{"header2", "value2"}}},
		{"Shorter2", args{h1: http.RawHeader{{"header1", "value1"}, {"header2", "value2"}}, h2: http.RawHeader{{"header1", "value1"}}}, http.RawHeader{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compareRawHeaders(tt.args.h1, tt.args.h2); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("compareRawHeaders() = %v, want %v", got, tt.want)
			}
		})
	}
}
