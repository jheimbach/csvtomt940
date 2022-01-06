package n26

import (
	"encoding/csv"
	"log"
	"os"
	"strings"

	"github.com/JHeimbach/csvtomt940/mt940"
	"github.com/Rhymond/go-money"
)

type N26 struct {
	Iban        string
	StartSaldo  int64
	HasCategory bool
	logger      *log.Logger
	data        *mt940.BankData
}

func New(iban string, startSaldo int64, hasCategory bool) *N26 {

	logger := log.New(os.Stdout, "[N26] ", log.Lmsgprefix)

	return &N26{
		logger:      logger,
		Iban:        iban,
		StartSaldo:  startSaldo,
		HasCategory: hasCategory,
	}
}

func (n *N26) ParseCsv(csvFile *os.File) *mt940.BankData {
	// extract banknumber and accountnumber from meta fields
	bankNumber, accountNumber := extractAccountAndBankNumber(n.Iban)

	n.data = &mt940.BankData{
		AccountNumber: accountNumber,
		BankNumber:    bankNumber,
	}

	// read rest of the file as csv
	cr := csv.NewReader(csvFile)
	cr.Comma = ','
	cr.LazyQuotes = true
	// header line
	_, err := cr.Read()
	if err != nil {
		n.logger.Fatalf("could not read data from csv %v", err)
	}

	transactions, err := cr.ReadAll()
	if err != nil {
		n.logger.Fatalf("could not read data from csv %v", err)
	}
	saldo := money.New(n.StartSaldo, "EUR")
	// create ingTransaction structs
	var ta = make([]mt940.Transaction, 0, len(transactions))
	for j, t := range transactions {
		ts, nSaldo, err := newTransactionFromCsv(t, saldo, n.HasCategory)
		if err != nil {
			log.Fatalf("could not convert entry to struct in line %d: %v", j, err)
		}
		saldo = nSaldo
		ta = append(ta, ts)
	}

	n.data.Transactions = ta

	return n.data
}

// extractAccountAndBankNumber returns blz and accountNumber from meta tags of the ING csv
func extractAccountAndBankNumber(iban string) (string, string) {
	// replace all whitespaces
	iban = strings.ReplaceAll(iban, " ", "")
	// blz begins in position 4 and has 8 chars
	// accountNumber begins in position 12 and has 10 chars (until the end of iban)
	return iban[4:12], strings.TrimSpace(iban[12:])
}
