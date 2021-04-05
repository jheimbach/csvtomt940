package n26

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
		name    string
		entry   []string
		want    *n26Transaction
		wantErr error
	}{
		{
			name:  "time is valid",
			entry: []string{"2000-01-02", "", "", "", "", "", "", "", "", ""},
			want: &n26Transaction{
				date:   time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				saldo:  money.New(0, "EUR"),
				amount: money.New(0, "EUR"),
			},
			wantErr: nil,
		},
		{
			name:    "time is invalid",
			entry:   []string{"2000-0102", "", "", "", "", "", "", "", "", ""},
			want:    nil,
			wantErr: fmt.Errorf("could not parse date from 2000-0102: %w", errors.New("parsing time \"2000-0102\" as \"2006-01-02\": cannot parse \"02\" as \"-\"")),
		},
		{
			name:  "both money values are valid",
			entry: []string{"2000-01-02", "", "", "", "", "", "12.00", "", "", ""},
			want: &n26Transaction{
				date:   time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				saldo:  money.New(1200, "EUR"),
				amount: money.New(1200, "EUR"),
			},
			wantErr: nil,
		},
		{
			name:    "amount money is invalid",
			entry:   []string{"2000-01-02", "", "", "", "", "", "12-00", "", "", ""},
			want:    nil,
			wantErr: fmt.Errorf("could not parse amount to int: %w", errors.New("strconv.Atoi: parsing \"12-00\": invalid syntax")),
		},
		{
			name:  "string fields are set",
			entry: []string{"2000-01-02", "test", "test2", "Income", "reference", "Salary", "12.00", "", "", ""},
			want: &n26Transaction{
				date:            time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				payee:           "test",
				transactionType: "Income",
				reference:       "reference",
				category:        "Salary",
				saldo:           money.New(1200, "EUR"),
				amount:          money.New(1200, "EUR"),
			},
			wantErr: nil,
		},
		{
			name:  "creditcard payment is credit",
			entry: []string{"2000-01-02", "test", "test2", "MasterCard Payment", "reference", "Salary", "12.00", "", "", ""},
			want: &n26Transaction{
				date:                  time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				payee:                 "test",
				transactionType:       "MasterCard Payment",
				transactionTypeLookup: "MasterCard Payment Credit",
				reference:             "reference",
				category:              "Salary",
				saldo:                 money.New(1200, "EUR"),
				amount:                money.New(1200, "EUR"),
			},
			wantErr: nil,
		},
		{
			name:  "creditcard payment is debit",
			entry: []string{"2000-01-02", "test", "test2", "MasterCard Payment", "reference", "Salary", "-12.00", "", "", ""},
			want: &n26Transaction{
				date:                  time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
				payee:                 "test",
				transactionType:       "MasterCard Payment",
				transactionTypeLookup: "MasterCard Payment Debit",
				reference:             "reference",
				category:              "Salary",
				saldo:                 money.New(-1200, "EUR"),
				amount:                money.New(-1200, "EUR"),
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := newTransactionFromCsv(tt.entry, money.New(0, "EUR"))
			if tt.wantErr != nil && err != nil {
				if tt.wantErr.Error() != err.Error() {
					t.Errorf("newTransactionFromCsv() got error:\n%v\n, wanted error:\n%v\n", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("newTransactionFromCsv() error = %v", err)
				return
			}
			ingTransactionsAreEqual(t, got, tt.want)
		})
	}
}

func Test_newTransactionFromCSV_CalculatesSaldoCorrect(t *testing.T) {
	tests := []struct {
		name       string
		entry      [][]string
		startSaldo *money.Money
		want       []*n26Transaction
	}{
		{
			name: "single transaction",
			entry: [][]string{
				{"2000-01-02", "test", "test2", "Income", "reference", "Salary", "12.00", "", "", ""},
			},
			startSaldo: money.New(0, "EUR"),
			want: []*n26Transaction{
				{
					date:            time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
					payee:           "test",
					transactionType: "Income",
					category:        "Salary",
					reference:       "reference",
					saldo:           money.New(1200, "EUR"),
					amount:          money.New(1200, "EUR"),
				},
			},
		},
		{
			name: "with non zero start saldo",
			entry: [][]string{
				{"2000-01-02", "test", "test2", "Income", "reference", "Salary", "12.00", "", "", ""},
			},
			startSaldo: money.New(1200, "EUR"),
			want: []*n26Transaction{
				{
					date:            time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
					payee:           "test",
					transactionType: "Income",
					category:        "Salary",
					reference:       "reference",
					saldo:           money.New(2400, "EUR"),
					amount:          money.New(1200, "EUR"),
				},
			},
		},
		{
			name: "with negative amount",
			entry: [][]string{
				{"2000-01-02", "test", "test2", "Income", "reference", "Salary", "-12.00", "", "", ""},
			},
			startSaldo: money.New(2400, "EUR"),
			want: []*n26Transaction{
				{
					date:            time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
					payee:           "test",
					transactionType: "Income",
					category:        "Salary",
					reference:       "reference",
					saldo:           money.New(1200, "EUR"),
					amount:          money.New(-1200, "EUR"),
				},
			},
		},
		{
			name: "multiple transaction",
			entry: [][]string{
				{"2000-01-02", "test", "test2", "Income", "reference", "Salary", "12.00", "", "", ""},
				{"2000-01-02", "test", "test2", "Income", "reference", "Salary", "12.00", "", "", ""},
			},
			startSaldo: money.New(0, "EUR"),
			want: []*n26Transaction{
				{
					date:            time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
					payee:           "test",
					transactionType: "Income",
					category:        "Salary",
					reference:       "reference",
					saldo:           money.New(1200, "EUR"),
					amount:          money.New(1200, "EUR"),
				},
				{
					date:            time.Date(2000, 01, 02, 00, 00, 00, 00, time.UTC),
					payee:           "test",
					transactionType: "Income",
					category:        "Salary",
					reference:       "reference",
					saldo:           money.New(2400, "EUR"),
					amount:          money.New(1200, "EUR"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var startSaldo = tt.startSaldo
			var transactions = make([]*n26Transaction, 0, len(tt.want))
			for _, entry := range tt.entry {
				got, gotSaldo, err := newTransactionFromCsv(entry, startSaldo)
				if err != nil {
					t.Errorf("newTransactionFromCsv() error = %v", err)
					return
				}
				startSaldo = gotSaldo
				transactions = append(transactions, got)
			}
			for i, want := range tt.want {
				ingTransactionsAreEqual(t, transactions[i], want)
			}
		})
	}
}

func ingTransactionsAreEqual(t *testing.T, a *n26Transaction, b *n26Transaction) {
	t.Helper()
	if !a.date.Equal(b.date) {
		t.Fatalf("date is not equal: %s !== %s", a.date.String(), b.date.String())
	}
	if a.payee != b.payee {
		t.Fatalf("payee is not equal: %s !== %s", a.payee, b.payee)
	}
	if a.transactionType != b.transactionType {
		t.Fatalf("transactionType is not equal: %s !== %s", a.transactionType, b.transactionType)
	}
	if a.category != b.category {
		t.Fatalf("category is not equal: %s !== %s", a.category, b.category)
	}
	if a.reference != b.reference {
		t.Fatalf("reference is not equal: %s !== %s", a.reference, b.reference)
	}
	if ok, _ := a.saldo.Equals(b.saldo); !ok {
		t.Fatalf("saldo is not equal: %s !== %s", a.saldo.Display(), b.saldo.Display())
	}
	if ok, _ := a.amount.Equals(b.amount); !ok {
		t.Fatalf("amount is not equal: %s !== %s", a.amount.Display(), b.amount.Display())
	}
}

func Test_ingtransaction_ConvertTOMT940(t1 *testing.T) {
	tests := []struct {
		name        string
		transaction *n26Transaction
		wantWriter  string
		wantErr     bool
	}{
		{
			name: "create mt940 for ingTransaction",
			transaction: &n26Transaction{
				date:                  time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				payee:                 "testname",
				transactionType:       "Überweisung",
				transactionTypeLookup: "Überweisung",
				reference:             "test",
				saldo:                 money.New(5000, "EUR"),
				amount:                money.New(-1050, "EUR"),
			},
			wantWriter: ":61:0001020102D10,50NTRFNONREF\r\n:86:020?00UEberweisung?20SVWZ+test?21KREF+NONREF?32testname\r\n",
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

func Test_ingtransaction_createSalesLine(t1 *testing.T) {
	tests := []struct {
		name        string
		transaction *n26Transaction
		wantWriter  string
		wantErr     bool
	}{
		{
			name: "create salesline with positive amount",
			transaction: &n26Transaction{
				date:            time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				payee:           "",
				transactionType: "",
				reference:       "",
				saldo:           nil,
				amount:          money.New(1050, "EUR"),
			},
			wantWriter: ":61:0001020102C10,50NTRFNONREF\r\n",
			wantErr:    false,
		},
		{
			name: "create salesline with negative amount",
			transaction: &n26Transaction{
				date:            time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				payee:           "",
				transactionType: "",
				reference:       "",
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
		transaction *n26Transaction
		wantWriter  string
		wantErr     bool
	}{
		{
			name: "empty reference line, empty auftraggeber",
			transaction: &n26Transaction{
				date:                  time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				payee:                 "",
				transactionType:       "MasterCard Payment",
				transactionTypeLookup: "MasterCard Payment Debit",
				reference:             "",
				saldo:                 nil,
				amount:                nil,
			},
			wantWriter: ":86:004?00MasterCard Payment?20KREF+NONREF\r\n",
			wantErr:    false,
		},
		{
			name: "reference line, empty auftraggeber",
			transaction: &n26Transaction{
				date:                  time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				payee:                 "",
				transactionType:       "MasterCard Payment",
				transactionTypeLookup: "MasterCard Payment Debit",
				reference:             "test",
				saldo:                 nil,
				amount:                nil,
			},
			wantWriter: ":86:004?00MasterCard Payment?20SVWZ+test?21KREF+NONREF\r\n",
			wantErr:    false,
		},
		{
			name: "reference line, with auftraggeber",
			transaction: &n26Transaction{
				date:                  time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				payee:                 "testname",
				transactionType:       "MasterCard Payment",
				transactionTypeLookup: "MasterCard Payment Debit",
				reference:             "test",
				saldo:                 nil,
				amount:                nil,
			},
			wantWriter: ":86:004?00MasterCard Payment?20SVWZ+test?21KREF+NONREF?32testname\r\n",
			wantErr:    false,
		},
		{
			name: "replaces transactionType umlauts",
			transaction: &n26Transaction{
				date:                  time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				payee:                 "testname",
				transactionType:       "Überweisung",
				transactionTypeLookup: "Überweisung",
				reference:             "test",
				saldo:                 nil,
				amount:                nil,
			},
			wantWriter: ":86:020?00UEberweisung?20SVWZ+test?21KREF+NONREF?32testname\r\n",
			wantErr:    false,
		},
		{
			name: "gvc code not found",
			transaction: &n26Transaction{
				date:            time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				payee:           "testname",
				transactionType: "Abschuss",
				reference:       "test",
				saldo:           nil,
				amount:          nil,
			},
			wantWriter: "",
			wantErr:    true,
		},
		{
			name: "reference line is too long",
			transaction: &n26Transaction{
				date:                  time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				payee:                 "testname",
				transactionType:       "MasterCard Payment",
				transactionTypeLookup: "MasterCard Payment Debit",
				reference:             strings.Repeat("a", 8*27),
				saldo:                 nil,
				amount:                nil,
			},
			wantWriter: "",
			wantErr:    true,
		},
		{
			name: "multipurpose line is too long",
			transaction: &n26Transaction{
				date:                  time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				payee:                 strings.Repeat("testname", 20),
				transactionType:       "MasterCard Payment",
				transactionTypeLookup: "MasterCard Payment Debit",
				reference:             strings.Repeat("a", 7*27),
				saldo:                 nil,
				amount:                nil,
			},
			wantWriter: "",
			wantErr:    true,
		},
		{
			name: "multipurpose line is split in parts",
			transaction: &n26Transaction{
				date:                  time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				payee:                 "testname",
				transactionType:       "MasterCard Payment",
				transactionTypeLookup: "MasterCard Payment Debit",
				reference:             strings.Repeat("a", 7*27),
				saldo:                 nil,
				amount:                nil,
			},
			wantWriter: fmt.Sprintf(":86:%s\r\n", strings.Join([]string{
				"004?00MasterCard Payment?20SVWZ+aaaaaaaaaaaaaaaaaaaaaa?21aaaaaaaa",
				"aaaaaaaaaaaaaaaaaaa?22aaaaaaaaaaaaaaaaaaaaaaaaaaa?23aaaaaaaaaaaaa",
				"aaaaaaaaaaaaaa?24aaaaaaaaaaaaaaaaaaaaaaaaaaa?25aaaaaaaaaaaaaaaaaa",
				"aaaaaaaaa?26aaaaaaaaaaaaaaaaaaaaaaaaaaa?27aaaaa?28KREF+NONREF?32t",
				"estname",
			}, "\r\n")),
			wantErr: false,
		},
		{
			name: "long payee name",
			transaction: &n26Transaction{
				date:                  time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC),
				payee:                 strings.Repeat("b", 53),
				transactionType:       "MasterCard Payment",
				transactionTypeLookup: "MasterCard Payment Debit",
				reference:             strings.Repeat("a", 27),
				saldo:                 nil,
				amount:                nil,
			},
			wantWriter: fmt.Sprintf(":86:%s\r\n", strings.Join([]string{
				"004?00MasterCard Payment?20SVWZ+aaaaaaaaaaaaaaaaaaaaaa?21aaaaa?22",
				"KREF+NONREF?32bbbbbbbbbbbbbbbbbbbbbbbbbbb?33bbbbbbbbbbbbbbbbbbbbb",
				"bbbbb",
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
