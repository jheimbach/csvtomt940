package ing

import (
	"bufio"
	"reflect"
	"strings"
	"testing"
)

func Test_getAccountNumber(t *testing.T) {
	type args struct {
		meta []string
	}
	tests := []struct {
		name          string
		args          args
		bankNumber    string
		accountNumber string
	}{
		{
			name: "without spaces",
			args: args{
				meta: []string{"", "", "IBAN;DE00111111110000000000"},
			},
			bankNumber:    "11111111",
			accountNumber: "0000000000",
		},
		{
			name: "with spaces",
			args: args{
				meta: []string{"", "", "IBAN;DE22 1111 1111 0000 0000 00"},
			},
			bankNumber:    "11111111",
			accountNumber: "0000000000",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bankNumber, accountNumber := getAccountNumber(tt.args.meta)
			if bankNumber != tt.bankNumber {
				t.Errorf("getAccountNumber() bankNumber = %v, accountNumber %v", bankNumber, tt.bankNumber)
			}
			if accountNumber != tt.accountNumber {
				t.Errorf("getAccountNumber() accountNumber = %v, accountNumber %v", accountNumber, tt.accountNumber)
			}
		})
	}
}

func Test_extractMetaFields(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:    "with breaklines",
			input:   strings.Repeat("1\n", 15),
			want:    []string{"1\n", "1\n", "1\n", "1\n", "1\n", "1\n", "1\n", "1\n", "1\n", "1\n", "1\n", "1\n", "1\n", "1\n"},
			wantErr: false,
		},
		{
			name:    "without breaklines",
			input:   strings.Repeat("1", 15),
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bufio.NewReader(strings.NewReader(tt.input))
			got, err := extractMetaFields(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractMetaFields() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractMetaFields() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_cleanUpTransactions(t *testing.T) {
	tests := []struct {
		name string
		ts   [][]string
		want [][]string
	}{
		{
			name: "test with 3",
			ts:   [][]string{{"1", "2"}, {"3", "4"}, {"5", "6"}},
			want: [][]string{{"5", "6"}, {"3", "4"}},
		},
		{
			name: "test with 4",
			ts:   [][]string{{"1", "2"}, {"3", "4"}, {"5", "6"}, {"7", "8"}},
			want: [][]string{{"7", "8"}, {"5", "6"}, {"3", "4"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanUpTransactions(tt.ts); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cleanUpTransactions() = %v, want %v", got, tt.want)
			}
		})
	}
}
