package main

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/Rhymond/go-money"
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
				t.Errorf("splitStringInParts() = %#v, wantBT %#v", got, tt.want)
			}
		})
	}
}

func Test_moneyStringToInt(t *testing.T) {
	type args struct {
		m string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "zero amount",
			args: args{
				m: "0",
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "empty amount",
			args: args{
				m: "",
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "zero cents",
			args: args{
				m: "12,00",
			},
			want:    1200,
			wantErr: false,
		},
		{
			name: "some cents",
			args: args{
				m: "12,12",
			},
			want:    1212,
			wantErr: false,
		},
		{
			name: "only cents",
			args: args{
				m: "0,12",
			},
			want:    12,
			wantErr: false,
		},
		{
			name: "thausend point",
			args: args{
				m: "1.000,12",
			},
			want:    100012,
			wantErr: false,
		},
		{
			name: "not a number",
			args: args{
				m: "adb",
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := moneyStringToInt(tt.args.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("moneyStringToInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("moneyStringToInt() got = %v, wantBT %v", got, tt.want)
			}
		})
	}
}

/*

func Test_parseMoneyValues(t *testing.T) {
	tests := []struct {
		name    string
		entry []string
		wantSaldo    *money.Money
		wantAmount   *money.Money
		wantErr error
	}{
		{
			name:       "both money values are valid",
			entry:      []string{"","","","","","12,00","EUR", "5,00", "EUR"},
			wantSaldo:  money.New(1200, "EUR"),
			wantAmount: money.New(500, "EUR"),
			wantErr:    nil,
		},
		{
			name:       "saldo money is invalid",
			entry:      []string{"","","","","","12-00","EUR", "5,00", "EUR"},
			wantSaldo:  money.New(1200, "EUR"),
			wantAmount: money.New(500, "EUR"),
			wantErr:    errors.New(""),
		},
		{
			name:       "amount money is invalid",
			entry:      []string{"","","","","","12,00","EUR", "5-00", "EUR"},
			wantSaldo:  money.New(1200, "EUR"),
			wantAmount: money.New(500, "EUR"),
			wantErr:    errors.New(""),
		},
	}
}*/

func Test_newTransactionFromCSV(t *testing.T) {
	tests := []struct {
		name    string
		entry   []string
		want    *transaction
		wantErr error
	}{
		{
			name:  "both times are valid",
			entry: []string{"02.01.2000", "03.02.2001", "", "", "", "", "", "", ""},
			want: &transaction{
				Buchung:          time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				Valuta:           time.Date(2001, 02, 03, 00, 00, 00, 00, time.UTC),
				Auftraggeber:     "",
				BuchungsText:     "",
				Verwendungszweck: "",
				Saldo:            money.New(0, ""),
				Amount:           money.New(0, ""),
			},
			wantErr: nil,
		},
		{
			name:  "bt times is invalid",
			entry: []string{"0201.2000", "03.02.2001", "", "", "", "", "", "", ""},
			want: &transaction{
				Buchung:          time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				Valuta:           time.Date(2001, 02, 03, 00, 00, 00, 00, time.UTC),
				Auftraggeber:     "",
				BuchungsText:     "",
				Verwendungszweck: "",
				Saldo:            money.New(0, ""),
				Amount:           money.New(0, ""),
			},
			wantErr: fmt.Errorf("could not parse Buchung: %w", errors.New("parsing time \"0201.2000\" as \"02.01.2006\": cannot parse \"01.2000\" as \".\"")),
		},
		{
			name:  "vt times is invalid",
			entry: []string{"02.01.2000", "0302.2001", "", "", "", "", "", "", ""},
			want: &transaction{
				Buchung:          time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				Valuta:           time.Date(2001, 02, 03, 00, 00, 00, 00, time.UTC),
				Auftraggeber:     "",
				BuchungsText:     "",
				Verwendungszweck: "",
				Saldo:            money.New(0, ""),
				Amount:           money.New(0, ""),
			},
			wantErr: fmt.Errorf("could not parse Valuta: %w", errors.New("parsing time \"0302.2001\" as \"02.01.2006\": cannot parse \"02.2001\" as \".\"")),
		},
		{
			name:  "both money values are valid",
			entry: []string{"02.01.2000", "02.01.2000", "", "", "", "12,00", "EUR", "5,00", "EUR"},
			want: &transaction{
				Buchung:          time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				Valuta:           time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				Auftraggeber:     "",
				BuchungsText:     "",
				Verwendungszweck: "",
				Saldo:            money.New(1200, "EUR"),
				Amount:           money.New(500, "EUR"),
			},
			wantErr: nil,
		},
		{
			name:  "saldo money is invalid",
			entry: []string{"02.01.2000", "02.01.2000", "", "", "", "12-00", "EUR", "5,00", "EUR"},
			want: &transaction{
				Buchung:          time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				Valuta:           time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				Auftraggeber:     "",
				BuchungsText:     "",
				Verwendungszweck: "",
				Saldo:            money.New(1200, "EUR"),
				Amount:           money.New(500, "EUR"),
			},
			wantErr: fmt.Errorf("could not parse saldo to int: %w", errors.New("strconv.Atoi: parsing \"12-00\": invalid syntax")),
		},
		{
			name:  "amount money is invalid",
			entry: []string{"02.01.2000", "02.01.2000", "", "", "", "12,00", "EUR", "5-00", "EUR"},
			want: &transaction{
				Buchung:          time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				Valuta:           time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				Auftraggeber:     "",
				BuchungsText:     "",
				Verwendungszweck: "",
				Saldo:            money.New(1200, "EUR"),
				Amount:           money.New(500, "EUR"),
			},
			wantErr: fmt.Errorf("could not parse amount to int: %w", errors.New("strconv.Atoi: parsing \"5-00\": invalid syntax")),
		},
		{
			name:  "string fields are set",
			entry: []string{"02.01.2000", "02.01.2000", "test", "test2", "test3", "12,00", "EUR", "5-00", "EUR"},
			want: &transaction{
				Buchung:          time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				Valuta:           time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				Auftraggeber:     "test",
				BuchungsText:     "test2",
				Verwendungszweck: "test3",
				Saldo:            money.New(1200, "EUR"),
				Amount:           money.New(500, "EUR"),
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
				t.Errorf("newTransactionFromCSV() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertUsageToFields(t *testing.T) {
	tests := []struct {
		name  string
		usage string
		want  string
	}{
		{
			name:  "empty usage",
			usage: "",
			want:  "?20KREF+NONREF",
		},
		{
			name:  "short usage, under 27 chars",
			usage: "this is a test",
			want:  "?20SVWZ+this is a test?21KREF+NONREF",
		},
		{
			name:  "long usage",
			usage: "VISA 4546 XXXX XXXX XXXX 1,75%AUSLANDSEINSATZENTGELT VISA CARD (DEBITKARTE) ARN24492150077637298081121\n",
			want:  "?20SVWZ+VISA 4546 XXXX XXXX XX?21XX 1,75%AUSLANDSEINSATZENTG?22ELT VISA CARD (DEBITKARTE)A?23RN24492150077637298081121?24KREF+NONREF",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertUsageToFields(tt.usage); got != tt.want {
				t.Errorf("convertUsageToFields() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Test_transaction_createSalesLine(t1 *testing.T) {
	tests := []struct {
		name        string
		transaction *transaction
		wantWriter  string
		wantErr     bool
	}{
		{
			name: "create salesline with positive amount",
			transaction: &transaction{
				Buchung:          time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				Valuta:           time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				Auftraggeber:     "",
				BuchungsText:     "",
				Verwendungszweck: "",
				Saldo:            nil,
				Amount:           money.New(1050, "EUR"),
			},
			wantWriter: ":61:0001020102C10,50NTRFNONREF\r\n",
			wantErr:    false,
		},
		{
			name: "create salesline with negative amount",
			transaction: &transaction{
				Buchung:          time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				Valuta:           time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				Auftraggeber:     "",
				BuchungsText:     "",
				Verwendungszweck: "",
				Saldo:            nil,
				Amount:           money.New(-1050, "EUR"),
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
				t1.Errorf("createSalesLine() gotWriter = %v, want %v", gotWriter, tt.wantWriter)
			}
		})
	}
}

func Test_transaction_createMultipurposeField(t1 *testing.T) {
	tests := []struct {
		name        string
		transaction *transaction
		wantWriter  string
		wantErr     bool
	}{
		{
			name: "empty usage line, empty auftraggeber",
			transaction: &transaction{
				Buchung:          time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				Valuta:           time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				Auftraggeber:     "",
				BuchungsText:     "Abschluss",
				Verwendungszweck: "",
				Saldo:            nil,
				Amount:           nil,
			},
			wantWriter: ":86:026?00Abschluss?20KREF+NONREF\r\n",
			wantErr:    false,
		},
		{
			name: "usage line, empty auftraggeber",
			transaction: &transaction{
				Buchung:          time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				Valuta:           time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				Auftraggeber:     "",
				BuchungsText:     "Abschluss",
				Verwendungszweck: "test",
				Saldo:            nil,
				Amount:           nil,
			},
			wantWriter: ":86:026?00Abschluss?20SVWZ+test?21KREF+NONREF\r\n",
			wantErr:    false,
		},
		{
			name: "usage line, with auftraggeber",
			transaction: &transaction{
				Buchung:          time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				Valuta:           time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				Auftraggeber:     "testname",
				BuchungsText:     "Abschluss",
				Verwendungszweck: "test",
				Saldo:            nil,
				Amount:           nil,
			},
			wantWriter: ":86:026?00Abschluss?20SVWZ+test?21KREF+NONREF?32testname\r\n",
			wantErr:    false,
		},
		{
			name: "gvc code not found",
			transaction: &transaction{
				Buchung:          time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				Valuta:           time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				Auftraggeber:     "testname",
				BuchungsText:     "Abschuss",
				Verwendungszweck: "test",
				Saldo:            nil,
				Amount:           nil,
			},
			wantWriter: "",
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			writer := &bytes.Buffer{}
			err := tt.transaction.createMultipurposeField(writer)
			if (err != nil) != tt.wantErr {
				t1.Errorf("createMultipurposeField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotWriter := writer.String(); gotWriter != tt.wantWriter {
				t1.Errorf("createMultipurposeField() gotWriter = %v, want %v", gotWriter, tt.wantWriter)
			}
		})
	}
}

func Test_transaction_convertToMt940(t1 *testing.T) {
	tests := []struct {
		name        string
		transaction *transaction
		wantWriter  string
		wantErr     bool
	}{
		{
			name: "create mt940 for transaction",
			transaction: &transaction{
				Buchung:          time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				Valuta:           time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				Auftraggeber:     "testname",
				BuchungsText:     "Abschluss",
				Verwendungszweck: "test",
				Saldo:            money.New(5000, "EUR"),
				Amount:           money.New(-1050, "EUR"),
			},
			wantWriter: ":61:0001020102D10,50NTRFNONREF\r\n:86:026?00Abschluss?20SVWZ+test?21KREF+NONREF?32testname\r\n",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			writer := &bytes.Buffer{}
			err := tt.transaction.convertToMt940(writer)
			if (err != nil) != tt.wantErr {
				t1.Errorf("convertToMt940() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotWriter := writer.String(); gotWriter != tt.wantWriter {
				t1.Errorf("convertToMt940() gotWriter = %v, want %v", gotWriter, tt.wantWriter)
			}
		})
	}
}
