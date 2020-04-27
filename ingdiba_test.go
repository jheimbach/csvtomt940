package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Rhymond/go-money"
)

func Test_newTransactionFromCSV(t *testing.T) {
	tests := []struct {
		name    string
		entry   []string
		want    *ingTransaction
		wantErr error
	}{
		{
			name:  "both times are valid",
			entry: []string{"02.01.2000", "03.02.2001", "", "", "", "", "", "", ""},
			want: &ingTransaction{
				date:            time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				valueDate:       time.Date(2001, 02, 03, 00, 00, 00, 00, time.UTC),
				client:          "",
				transactionType: "",
				usage:           "",
				saldo:           money.New(0, ""),
				amount:          money.New(0, ""),
			},
			wantErr: nil,
		},
		{
			name:  "bt times is invalid",
			entry: []string{"0201.2000", "03.02.2001", "", "", "", "", "", "", ""},
			want: &ingTransaction{
				date:            time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				valueDate:       time.Date(2001, 02, 03, 00, 00, 00, 00, time.UTC),
				client:          "",
				transactionType: "",
				usage:           "",
				saldo:           money.New(0, ""),
				amount:          money.New(0, ""),
			},
			wantErr: fmt.Errorf("could not parse date: %w", errors.New("parsing time \"0201.2000\" as \"02.01.2006\": cannot parse \"01.2000\" as \".\"")),
		},
		{
			name:  "vt times is invalid",
			entry: []string{"02.01.2000", "0302.2001", "", "", "", "", "", "", ""},
			want: &ingTransaction{
				date:            time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				valueDate:       time.Date(2001, 02, 03, 00, 00, 00, 00, time.UTC),
				client:          "",
				transactionType: "",
				usage:           "",
				saldo:           money.New(0, ""),
				amount:          money.New(0, ""),
			},
			wantErr: fmt.Errorf("could not parse valueDate: %w", errors.New("parsing time \"0302.2001\" as \"02.01.2006\": cannot parse \"02.2001\" as \".\"")),
		},
		{
			name:  "both money values are valid",
			entry: []string{"02.01.2000", "02.01.2000", "", "", "", "12,00", "EUR", "5,00", "EUR"},
			want: &ingTransaction{
				date:            time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				valueDate:       time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				client:          "",
				transactionType: "",
				usage:           "",
				saldo:           money.New(1200, "EUR"),
				amount:          money.New(500, "EUR"),
			},
			wantErr: nil,
		},
		{
			name:  "saldo money is invalid",
			entry: []string{"02.01.2000", "02.01.2000", "", "", "", "12-00", "EUR", "5,00", "EUR"},
			want: &ingTransaction{
				date:            time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				valueDate:       time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				client:          "",
				transactionType: "",
				usage:           "",
				saldo:           money.New(1200, "EUR"),
				amount:          money.New(500, "EUR"),
			},
			wantErr: fmt.Errorf("could not parse saldo to int: %w", errors.New("strconv.Atoi: parsing \"12-00\": invalid syntax")),
		},
		{
			name:  "amount money is invalid",
			entry: []string{"02.01.2000", "02.01.2000", "", "", "", "12,00", "EUR", "5-00", "EUR"},
			want: &ingTransaction{
				date:            time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				valueDate:       time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				client:          "",
				transactionType: "",
				usage:           "",
				saldo:           money.New(1200, "EUR"),
				amount:          money.New(500, "EUR"),
			},
			wantErr: fmt.Errorf("could not parse amount to int: %w", errors.New("strconv.Atoi: parsing \"5-00\": invalid syntax")),
		},
		{
			name:  "string fields are set",
			entry: []string{"02.01.2000", "02.01.2000", "test", "test2", "test3", "12,00", "EUR", "5-00", "EUR"},
			want: &ingTransaction{
				date:            time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				valueDate:       time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				client:          "test",
				transactionType: "test2",
				usage:           "test3",
				saldo:           money.New(1200, "EUR"),
				amount:          money.New(500, "EUR"),
			},
			wantErr: fmt.Errorf("could not parse amount to int: %w", errors.New("strconv.Atoi: parsing \"5-00\": invalid syntax")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newTransactionFromCSV(tt.entry)
			if tt.wantErr != nil && err != nil {
				if tt.wantErr.Error() != err.Error() {
					t.Errorf("moneyStringToInt() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("moneyStringToInt() error = %v", err)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newTransactionFromCSV() got = %v, wantWriter %v", got, tt.want)
			}
		})
	}
}

func Test_ingtransaction_createSalesLine(t1 *testing.T) {
	tests := []struct {
		name        string
		transaction *ingTransaction
		wantWriter  string
		wantErr     bool
	}{
		{
			name: "create salesline with positive amount",
			transaction: &ingTransaction{
				date:            time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				valueDate:       time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				client:          "",
				transactionType: "",
				usage:           "",
				saldo:           nil,
				amount:          money.New(1050, "EUR"),
			},
			wantWriter: ":61:0001020102C10,50NTRFNONREF\r\n",
			wantErr:    false,
		},
		{
			name: "create salesline with negative amount",
			transaction: &ingTransaction{
				date:            time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				valueDate:       time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				client:          "",
				transactionType: "",
				usage:           "",
				saldo:           nil,
				amount:          money.New(-1050, "EUR"),
			},
			wantWriter: ":61:0001020102D10,50NTRFNONREF\r\n",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			writer := &bytes.Buffer{}
			err := tt.transaction.createSalesLine(writer)
			if (err != nil) != tt.wantErr {
				t1.Errorf("createSalesLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotWriter := writer.String(); gotWriter != tt.wantWriter {
				t1.Errorf("createSalesLine() gotWriter = %v, wantWriter %v", gotWriter, tt.wantWriter)
			}
		})
	}
}

func Test_ingtransaction_createMultipurposeField(t1 *testing.T) {
	tests := []struct {
		name        string
		transaction *ingTransaction
		wantWriter  string
		wantErr     bool
	}{
		{
			name: "empty usage line, empty auftraggeber",
			transaction: &ingTransaction{
				date:            time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				valueDate:       time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				client:          "",
				transactionType: "Lastschrift",
				usage:           "",
				saldo:           nil,
				amount:          nil,
			},
			wantWriter: ":86:005?00Lastschrift?20KREF+NONREF\r\n",
			wantErr:    false,
		},
		{
			name: "usage line, empty auftraggeber",
			transaction: &ingTransaction{
				date:            time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				valueDate:       time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				client:          "",
				transactionType: "Lastschrift",
				usage:           "test",
				saldo:           nil,
				amount:          nil,
			},
			wantWriter: ":86:005?00Lastschrift?20SVWZ+test?21KREF+NONREF\r\n",
			wantErr:    false,
		},
		{
			name: "usage line, with auftraggeber",
			transaction: &ingTransaction{
				date:            time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				valueDate:       time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				client:          "testname",
				transactionType: "Lastschrift",
				usage:           "test",
				saldo:           nil,
				amount:          nil,
			},
			wantWriter: ":86:005?00Lastschrift?20SVWZ+test?21KREF+NONREF?32testname\r\n",
			wantErr:    false,
		},
		{
			name: "replaces transactionType umlauts",
			transaction: &ingTransaction{
				date:            time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				valueDate:       time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				client:          "testname",
				transactionType: "Überweisung",
				usage:           "test",
				saldo:           nil,
				amount:          nil,
			},
			wantWriter: ":86:020?00UEberweisung?20SVWZ+test?21KREF+NONREF?32testname\r\n",
			wantErr:    false,
		},
		{
			name: "gvc code not found",
			transaction: &ingTransaction{
				date:            time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				valueDate:       time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				client:          "testname",
				transactionType: "Abschuss",
				usage:           "test",
				saldo:           nil,
				amount:          nil,
			},
			wantWriter: "",
			wantErr:    true,
		},
		{
			name: "usage line is too long",
			transaction: &ingTransaction{
				date:            time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				valueDate:       time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				client:          "testname",
				transactionType: "Lastschrift",
				usage:           strings.Repeat("a", 8*27),
				saldo:           nil,
				amount:          nil,
			},
			wantWriter: "",
			wantErr:    true,
		},
		{
			name: "multipurpose line is too long",
			transaction: &ingTransaction{
				date:            time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				valueDate:       time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				client:          strings.Repeat("testname", 20),
				transactionType: "Lastschrift",
				usage:           strings.Repeat("a", 7*27),
				saldo:           nil,
				amount:          nil,
			},
			wantWriter: "",
			wantErr:    true,
		},
		{
			name: "multipurpose line is split in parts",
			transaction: &ingTransaction{
				date:            time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				valueDate:       time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				client:          "testname",
				transactionType: "Lastschrift",
				usage:           strings.Repeat("a", 7*27),
				saldo:           nil,
				amount:          nil,
			},
			wantWriter: fmt.Sprintf(":86:%s\r\n", strings.Join([]string{
				"005?00Lastschrift?20SVWZ+aaaaaaaaaaaaaaaaaaaaaa?21aaaaaaaaaaaaaaa",
				"aaaaaaaaaaaa?22aaaaaaaaaaaaaaaaaaaaaaaaaaa?23aaaaaaaaaaaaaaaaaaaa",
				"aaaaaaa?24aaaaaaaaaaaaaaaaaaaaaaaaaaa?25aaaaaaaaaaaaaaaaaaaaaaaaa",
				"aa?26aaaaaaaaaaaaaaaaaaaaaaaaaaa?27aaaaa?28KREF+NONREF?32testname",
			}, "\r\n")),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			writer := &bytes.Buffer{}
			err := tt.transaction.createMultipurposeLine(writer)
			if (err != nil) != tt.wantErr {
				t1.Errorf("createMultipurposeLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotWriter := writer.String(); gotWriter != tt.wantWriter {
				t1.Errorf("createMultipurposeLine() gotWriter = %#v, wantWriter %#v", gotWriter, tt.wantWriter)
			}
		})
	}
}

func Test_ingtransaction_ConvertTOMT940(t1 *testing.T) {
	tests := []struct {
		name        string
		transaction *ingTransaction
		wantWriter  string
		wantErr     bool
	}{
		{
			name: "create mt940 for ingTransaction",
			transaction: &ingTransaction{
				date:            time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				valueDate:       time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				client:          "testname",
				transactionType: "Abschluss",
				usage:           "test",
				saldo:           money.New(5000, "EUR"),
				amount:          money.New(-1050, "EUR"),
			},
			wantWriter: ":61:0001020102D10,50NTRFNONREF\r\n:86:805?00Abschluss?20SVWZ+test?21KREF+NONREF?32testname\r\n",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			writer := &bytes.Buffer{}
			err := tt.transaction.ConvertToMT940(writer)
			if (err != nil) != tt.wantErr {
				t1.Errorf("ConvertToMT940() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotWriter := writer.String(); gotWriter != tt.wantWriter {
				t1.Errorf("ConvertToMT940() gotWriter = %#v, wantWriter %#v", gotWriter, tt.wantWriter)
			}
		})
	}
}

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

func Test_cleanUpTransactions(t *testing.T) {
	type args struct {
		ts [][]string
	}
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
