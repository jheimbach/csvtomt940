package mt940

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/Rhymond/go-money"
)

type mockTransaction struct {
	convert func(writer io.Writer) error
	saldo   func() *money.Money
	amount  func() *money.Money
	date    func() time.Time
}

func (m *mockTransaction) ConvertToMT940(writer io.Writer) error {
	return m.convert(writer)
}

func (m *mockTransaction) Saldo() *money.Money {
	return m.saldo()
}

func (m *mockTransaction) Amount() *money.Money {
	return m.amount()
}

func (m *mockTransaction) Date() time.Time {
	return m.date()
}

func Test_SwiftTransactions_ConvertToMT940(t *testing.T) {
	type fields struct {
		accountNumber string
		bankNumber    string
		transactions  []Transaction
	}
	tests := []struct {
		name    string
		fields  fields
		wantW   string
		wantErr bool
	}{
		{
			name: "empty struct",
			fields: fields{
				accountNumber: "",
				bankNumber:    "",
				transactions:  nil,
			},
			wantW:   ":20:CSVTOMT940\r\n",
			wantErr: true,
		},
		{
			name: "filled bank and account number",
			fields: fields{
				accountNumber: "0000000000",
				bankNumber:    "11111111",
				transactions:  nil,
			},
			wantW:   ":20:CSVTOMT940\r\n:25:11111111/0000000000\r\n:28C:0\r\n",
			wantErr: true,
		},
		{
			name: "with single transaction",
			fields: fields{
				accountNumber: "0000000000",
				bankNumber:    "11111111",
				transactions: []Transaction{&mockTransaction{
					convert: func(writer io.Writer) error {
						return nil
					},
					saldo: func() *money.Money {
						return money.New(10000, "EUR")
					},
					amount: func() *money.Money {
						return money.New(100, "EUR")
					},
					date: func() time.Time {
						return time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)
					},
				}},
			},
			wantW:   ":20:CSVTOMT940\r\n:25:11111111/0000000000\r\n:28C:0\r\n:60F:C000102EUR99,00\r\n:62F:C000102EUR100,00\r\n",
			wantErr: false,
		},
		{
			name: "two transactions",
			fields: fields{
				accountNumber: "0000000000",
				bankNumber:    "11111111",
				transactions: []Transaction{&mockTransaction{
					convert: func(writer io.Writer) error {
						return nil
					},
					saldo: func() *money.Money {
						return money.New(10000, "EUR")
					},
					amount: func() *money.Money {
						return money.New(100, "EUR")
					},
					date: func() time.Time {
						return time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)
					},
				}, &mockTransaction{
					convert: func(writer io.Writer) error {
						return nil
					},
					saldo: func() *money.Money {
						return money.New(5000, "EUR")
					},
					amount: func() *money.Money {
						return money.New(100, "EUR")
					},
					date: func() time.Time {
						return time.Date(2001, 2, 3, 0, 0, 0, 0, time.UTC)
					},
				}},
			},
			wantW:   ":20:CSVTOMT940\r\n:25:11111111/0000000000\r\n:28C:0\r\n:60F:C000102EUR99,00\r\n:62F:C010203EUR50,00\r\n",
			wantErr: false,
		},
		{
			name: "transaction with convert",
			fields: fields{
				accountNumber: "0000000000",
				bankNumber:    "11111111",
				transactions: []Transaction{&mockTransaction{
					convert: func(writer io.Writer) error {
						_, _ = writer.Write([]byte(":61:0001020102D10,50NTRFNONREF\r\n:86:026?00Abschluss?20KREF+NONREF\r\n"))
						return nil
					},
					saldo: func() *money.Money {
						return money.New(10000, "EUR")
					},
					amount: func() *money.Money {
						return money.New(-1050, "EUR")
					},
					date: func() time.Time {
						return time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)
					},
				}, &mockTransaction{
					convert: func(writer io.Writer) error {
						_, _ = writer.Write([]byte(":61:0001020102C1,00NTRFNONREF\r\n:86:026?00Abschluss?20KREF+NONREF\r\n"))
						return nil
					},
					saldo: func() *money.Money {
						return money.New(5000, "EUR")
					},
					amount: func() *money.Money {
						return money.New(100, "EUR")
					},
					date: func() time.Time {
						return time.Date(2001, 2, 3, 0, 0, 0, 0, time.UTC)
					},
				}},
			},
			wantW:   ":20:CSVTOMT940\r\n:25:11111111/0000000000\r\n:28C:0\r\n:60F:C000102EUR110,50\r\n:61:0001020102D10,50NTRFNONREF\r\n:86:026?00Abschluss?20KREF+NONREF\r\n:61:0001020102C1,00NTRFNONREF\r\n:86:026?00Abschluss?20KREF+NONREF\r\n:62F:C010203EUR50,00\r\n",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &BankData{
				AccountNumber: tt.fields.accountNumber,
				BankNumber:    tt.fields.bankNumber,
				Transactions:  tt.fields.transactions,
			}
			w := &bytes.Buffer{}
			err := s.ConvertToMT940(w)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToMT940() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("ConvertToMT940() gotW = %v, accountNumber %v", gotW, tt.wantW)
			}
		})
	}
}

func Test_SwiftTransactions_createAccountLine(t *testing.T) {
	type fields struct {
		accountNumber string
		bankNumber    string
		transactions  []Transaction
	}
	tests := []struct {
		name       string
		fields     fields
		wantWriter string
		wantErr    bool
	}{
		{
			name: "empty struct",
			fields: fields{
				accountNumber: "",
				bankNumber:    "",
				transactions:  nil,
			},
			wantWriter: "",
			wantErr:    true,
		},
		{
			name: "bankNumber is empty",
			fields: fields{
				accountNumber: "0000000000",
				bankNumber:    "",
				transactions:  nil,
			},
			wantWriter: "",
			wantErr:    true,
		},
		{
			name: "accountNumber is empty",
			fields: fields{
				accountNumber: "",
				bankNumber:    "00000000",
				transactions:  nil,
			},
			wantWriter: "",
			wantErr:    true,
		},
		{
			name: "bank and account number is filled",
			fields: fields{
				accountNumber: "0000000000",
				bankNumber:    "11111111",
				transactions:  nil,
			},
			wantWriter: ":25:11111111/0000000000\r\n",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &BankData{
				AccountNumber: tt.fields.accountNumber,
				BankNumber:    tt.fields.bankNumber,
				Transactions:  tt.fields.transactions,
			}
			writer := &bytes.Buffer{}
			err := s.createAccountLine(writer)
			if (err != nil) != tt.wantErr {
				t.Errorf("createAccountLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotWriter := writer.String(); gotWriter != tt.wantWriter {
				t.Errorf("createAccountLine() gotWriter = %v, accountNumber %v", gotWriter, tt.wantWriter)
			}
		})
	}
}

func Test_SwiftTransactions_createBeginSaldoLine(t *testing.T) {
	type fields struct {
		accountNumber string
		bankNumber    string
		transactions  []Transaction
	}
	tests := []struct {
		name       string
		fields     fields
		wantWriter string
		wantErr    bool
	}{
		{
			name: "empty struct",
			fields: fields{
				accountNumber: "",
				bankNumber:    "",
				transactions:  nil,
			},
			wantWriter: "",
			wantErr:    true,
		},
		{
			name: "positive saldo, positive amount",
			fields: fields{
				accountNumber: "",
				bankNumber:    "",
				transactions: []Transaction{
					&mockTransaction{
						convert: nil,
						saldo: func() *money.Money {
							return money.New(10000, "EUR")
						},
						amount: func() *money.Money {
							return money.New(100, "EUR")
						},
						date: func() time.Time {
							return time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC)
						},
					},
				},
			},
			wantWriter: ":60F:C000102EUR99,00\r\n",
			wantErr:    false,
		},
		{
			name: "positive saldo, negative amount",
			fields: fields{
				accountNumber: "",
				bankNumber:    "",
				transactions: []Transaction{
					&mockTransaction{
						convert: nil,
						saldo: func() *money.Money {
							return money.New(10000, "EUR")
						},
						amount: func() *money.Money {
							return money.New(-100, "EUR")
						},
						date: func() time.Time {
							return time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC)
						},
					},
				},
			},
			wantWriter: ":60F:C000102EUR101,00\r\n",
			wantErr:    false,
		},
		{
			name: "negative saldo, postitive amount",
			fields: fields{
				accountNumber: "",
				bankNumber:    "",
				transactions: []Transaction{
					&mockTransaction{
						convert: nil,
						saldo: func() *money.Money {
							return money.New(-10000, "EUR")
						},
						amount: func() *money.Money {
							return money.New(100, "EUR")
						},
						date: func() time.Time {
							return time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC)
						},
					},
				},
			},
			wantWriter: ":60F:D000102EUR101,00\r\n",
			wantErr:    false,
		},
		{
			name: "negative saldo, negative amount",
			fields: fields{
				accountNumber: "",
				bankNumber:    "",
				transactions: []Transaction{
					&mockTransaction{
						convert: nil,
						saldo: func() *money.Money {
							return money.New(-10000, "EUR")
						},
						amount: func() *money.Money {
							return money.New(-100, "EUR")
						},
						date: func() time.Time {
							return time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC)
						},
					},
				},
			},
			wantWriter: ":60F:D000102EUR99,00\r\n",
			wantErr:    false,
		},
		{
			name: "different currency between saldo and amount",
			fields: fields{
				accountNumber: "",
				bankNumber:    "",
				transactions: []Transaction{
					&mockTransaction{
						convert: nil,
						saldo: func() *money.Money {
							return money.New(-10000, "EUR")
						},
						amount: func() *money.Money {
							return money.New(-100, "USD")
						},
						date: func() time.Time {
							return time.Date(2000, 01, 02, 0, 0, 0, 0, time.UTC)
						},
					},
				},
			},
			wantWriter: "",
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &BankData{
				AccountNumber: tt.fields.accountNumber,
				BankNumber:    tt.fields.bankNumber,
				Transactions:  tt.fields.transactions,
			}
			writer := &bytes.Buffer{}
			err := s.createStartSaldoLine(writer)
			if (err != nil) != tt.wantErr {
				t.Errorf("createStartSaldoLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotWriter := writer.String(); gotWriter != tt.wantWriter {
				t.Errorf("createStartSaldoLine() gotWriter = %v, accountNumber %v", gotWriter, tt.wantWriter)
			}
		})
	}
}

func Test_SwiftTransactions_createEndSaldoLine(t *testing.T) {
	type fields struct {
		accountNumber string
		bankNumber    string
		transactions  []Transaction
	}
	tests := []struct {
		name       string
		fields     fields
		wantWriter string
		wantErr    bool
	}{
		{
			name: "empty struct",
			fields: fields{
				accountNumber: "",
				bankNumber:    "",
				transactions:  nil,
			},
			wantWriter: "",
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &BankData{
				AccountNumber: tt.fields.accountNumber,
				BankNumber:    tt.fields.bankNumber,
				Transactions:  tt.fields.transactions,
			}
			writer := &bytes.Buffer{}
			err := s.createEndSaldoLine(writer)
			if (err != nil) != tt.wantErr {
				t.Errorf("createEndSaldoLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotWriter := writer.String(); gotWriter != tt.wantWriter {
				t.Errorf("createEndSaldoLine() gotWriter = %v, accountNumber %v", gotWriter, tt.wantWriter)
			}
		})
	}
}

func Test_SwiftTransactions_createHeaderLine(t *testing.T) {
	type fields struct {
		accountNumber string
		bankNumber    string
		transactions  []Transaction
	}
	tests := []struct {
		name       string
		fields     fields
		wantWriter string
		wantErr    bool
	}{
		{
			name: "create header line",
			fields: fields{
				accountNumber: "",
				bankNumber:    "",
				transactions:  nil,
			},
			wantWriter: ":20:CSVTOMT940\r\n",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &BankData{
				AccountNumber: tt.fields.accountNumber,
				BankNumber:    tt.fields.bankNumber,
				Transactions:  tt.fields.transactions,
			}
			writer := &bytes.Buffer{}
			err := s.createHeaderLine(writer)
			if (err != nil) != tt.wantErr {
				t.Errorf("createHeaderLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotWriter := writer.String(); gotWriter != tt.wantWriter {
				t.Errorf("createHeaderLine() gotWriter = %v, accountNumber %v", gotWriter, tt.wantWriter)
			}
		})
	}
}

func Test_SwiftTransactions_createStatementLine(t *testing.T) {
	type fields struct {
		accountNumber string
		bankNumber    string
		transactions  []Transaction
	}
	tests := []struct {
		name       string
		fields     fields
		wantWriter string
		wantErr    bool
	}{
		{
			name: "create statement line",
			fields: fields{
				accountNumber: "",
				bankNumber:    "",
				transactions:  nil,
			},
			wantWriter: ":28C:0\r\n",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &BankData{
				AccountNumber: tt.fields.accountNumber,
				BankNumber:    tt.fields.bankNumber,
				Transactions:  tt.fields.transactions,
			}
			writer := &bytes.Buffer{}
			err := s.createStatementLine(writer)
			if (err != nil) != tt.wantErr {
				t.Errorf("createStatementLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotWriter := writer.String(); gotWriter != tt.wantWriter {
				t.Errorf("createStatementLine() gotWriter = %v, accountNumber %v", gotWriter, tt.wantWriter)
			}
		})
	}
}
