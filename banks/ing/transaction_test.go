package ing

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Rhymond/go-money"
)

func Test_newTransactionFromCSV(t *testing.T) {
	tests := []struct {
		name        string
		entry       []string
		hasCategory bool
		want        *ingTransaction
		wantErr     error
	}{
		{
			name:  "both times are valid",
			entry: []string{"02.01.2000", "03.02.2001", "", "", "", "", "", "", ""},
			want: &ingTransaction{
				date:      time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				valueDate: time.Date(2001, 02, 03, 00, 00, 00, 00, time.UTC),
				saldo:     money.New(0, ""),
				amount:    money.New(0, ""),
			},
			wantErr: nil,
		},
		{
			name:    "bt times is invalid",
			entry:   []string{"0201.2000", "03.02.2001", "", "", "", "", "", "", ""},
			want:    nil,
			wantErr: fmt.Errorf("could not parse date: %w", errors.New("parsing time \"0201.2000\" as \"02.01.2006\": cannot parse \"01.2000\" as \".\"")),
		},
		{
			name:    "vt times is invalid",
			entry:   []string{"02.01.2000", "0302.2001", "", "", "", "", "", "", ""},
			want:    nil,
			wantErr: fmt.Errorf("could not parse valueDate: %w", errors.New("parsing time \"0302.2001\" as \"02.01.2006\": cannot parse \"02.2001\" as \".\"")),
		},
		{
			name:  "both money values are valid",
			entry: []string{"02.01.2000", "02.01.2000", "", "", "", "12,00", "EUR", "5,00", "EUR"},
			want: &ingTransaction{
				date:      time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				valueDate: time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				saldo:     money.New(1200, "EUR"),
				amount:    money.New(500, "EUR"),
			},
			wantErr: nil,
		},
		{
			name:    "saldo money is invalid",
			entry:   []string{"02.01.2000", "02.01.2000", "", "", "", "12-00", "EUR", "5,00", "EUR"},
			want:    nil,
			wantErr: fmt.Errorf("could not parse saldo to int: %w", errors.New("strconv.Atoi: parsing \"12-00\": invalid syntax")),
		},
		{
			name:    "amount money is invalid",
			entry:   []string{"02.01.2000", "02.01.2000", "", "", "", "12,00", "EUR", "5-00", "EUR"},
			want:    nil,
			wantErr: fmt.Errorf("could not parse amount to int: %w", errors.New("strconv.Atoi: parsing \"5-00\": invalid syntax")),
		},
		{
			name:  "string fields are set",
			entry: []string{"02.01.2000", "02.01.2000", "test", "test2", "test3", "12,00", "EUR", "5,00", "EUR"},
			want: &ingTransaction{
				date:            time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				valueDate:       time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				client:          "test",
				transactionType: "test2",
				usage:           "test3",
				saldo:           money.New(1200, "EUR"),
				amount:          money.New(500, "EUR"),
			},
			wantErr: nil,
		},
		{
			hasCategory: true,
			name:        "string fields are set and transaction has category",
			entry:       []string{"02.01.2000", "02.01.2000", "client", "transactionType", "category", "usage", "12,00", "EUR", "5,00", "EUR"},
			want: &ingTransaction{
				date:            time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				valueDate:       time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				client:          "client",
				transactionType: "transactionType",
				category:        "category",
				usage:           "usage",
				saldo:           money.New(1200, "EUR"),
				amount:          money.New(500, "EUR"),
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newTransactionFromCSV(tt.entry, tt.hasCategory)
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
			ingTransactionsAreEqual(t, got, tt.want)
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
		{
			name: "create salesline with different dates",
			transaction: &ingTransaction{
				date:            time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				valueDate:       time.Date(2000, 01, 01, 0, 0, 0, 0, time.UTC),
				client:          "",
				transactionType: "",
				usage:           "",
				saldo:           nil,
				amount:          money.New(-1050, "EUR"),
			},
			wantWriter: ":61:0001010102D10,50NTRFNONREF\r\n",
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
		{
			name: "long client name",
			transaction: &ingTransaction{
				date:            time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				valueDate:       time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				client:          strings.Repeat("b", 53),
				transactionType: "Lastschrift",
				usage:           strings.Repeat("a", 27),
				saldo:           nil,
				amount:          nil,
			},
			wantWriter: fmt.Sprintf(":86:%s\r\n", strings.Join([]string{
				"005?00Lastschrift?20SVWZ+aaaaaaaaaaaaaaaaaaaaaa?21aaaaa?22KREF+NO",
				"NREF?32bbbbbbbbbbbbbbbbbbbbbbbbbbb?33bbbbbbbbbbbbbbbbbbbbbbbbbb",
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

//":86:005?00Lastschrift?20SVWZ+aaaaaaaaaaaaaaaaaaaaaa?21aaaaa?22KREF+NO\r\nNREF?32bbbbbbbbbbbbbbbbbbbbbbbbbbb?33bbbbbbbbbbbbbbbbbbbbbbbbbb\r\n"
//":86:005?00Lastschrift?20SVWZ+aaaaaaaaaaaaaaaaaaaaaa?21aaaaa?28KREF+NO\r\nNREF?32bbbbbbbbbbbbbbbbbbbbbbbbbbb?33bbbbbbbbbbbbbbbbbbbbbbbbbb\r\n"
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

func ingTransactionsAreEqual(t *testing.T, a *ingTransaction, b *ingTransaction) {
	t.Helper()
	if !a.date.Equal(b.date) {
		t.Fatalf("date is not equal: %s !== %s", a.date.String(), b.date.String())
	}
	if !a.valueDate.Equal(b.valueDate) {
		t.Fatalf("valueDate is not equal: %s !== %s", a.valueDate.String(), b.valueDate.String())
	}
	if a.client != b.client {
		t.Fatalf("client is not equal: %s !== %s", a.client, b.client)
	}
	if a.transactionType != b.transactionType {
		t.Fatalf("transactionType is not equal: %s !== %s", a.transactionType, b.transactionType)
	}
	if a.category != b.category {
		t.Fatalf("category is not equal: %s !== %s", a.category, b.category)
	}
	if a.usage != b.usage {
		t.Fatalf("usage is not equal: %s !== %s", a.usage, b.usage)
	}
	if ok, _ := a.saldo.Equals(b.saldo); !ok {
		t.Fatalf("saldo is not equal: %s !== %s", a.saldo.Display(), b.saldo.Display())
	}
	if ok, _ := a.amount.Equals(b.amount); !ok {
		t.Fatalf("amount is not equal: %s !== %s", a.amount.Display(), b.amount.Display())
	}
}
