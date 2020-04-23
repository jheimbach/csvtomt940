package main

import (
	"reflect"
	"testing"
)

func Test_splitStringInParts(t *testing.T) {
	type args struct {
		s string
		l int
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "each char is own string",
			args: args{
				s: "abc",
				l: 1,
			},
			want: []string{"a", "b", "c"},
		},
		{
			name: "split after 3rd char",
			args: args{
				s: "abcabcabc",
				l: 3,
			},
			want: []string{"abc", "abc", "abc"},
		},
		{
			name: "split after 27th char",
			args: args{
				s: "SVWZ+NR7778648141 INTERNET KAUFUMSATZ 25.12 256515 ARN85941831134325711900635",
				l: 27,
			},
			want: []string{"SVWZ+NR7778648141 INTERNETK", "AUFUMSATZ 25.12 256515 ARN8", "5941831134325711900635"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := splitStringInParts(tt.args.s, tt.args.l); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitStringInParts() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
