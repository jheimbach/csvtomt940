package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Rhymond/go-money"
)

// this consts are used to determine the order of fields in the ing-diba csv
const (
	date int = iota
	valueDate
	client
	transactionType
	usageLine
	saldo
	sCurrency
	amount
	aCurrency
)

// GVCCodes returns the GVC Code for the given transactionType, note this list is not complete, other values are possible
var GVCCodes = map[string]string{
	"Abschluss":                   "026",
	"Gutschrift aus Dauerauftrag": "052",
	"Lastschrift":                 "005",
	"Gutschrift":                  "051",
	"Überweisung":                 "020",
	"Entgelt":                     "027",
	"Retouren":                    "059",
}

// ingTransaction is the implementation of Transaction for ing-diba csv format
type ingTransaction struct {
	date            time.Time
	valueDate       time.Time
	client          string
	transactionType string
	usage           string
	saldo           *money.Money
	amount          *money.Money
}

// Saldo returns saldo field
func (t *ingTransaction) Saldo() *money.Money {
	return t.saldo
}

// Amount returns amount field
func (t *ingTransaction) Amount() *money.Money {
	return t.amount
}

// Date returns date field
func (t *ingTransaction) Date() time.Time {
	return t.date
}

// newTransactionFromCSV returns a transaction from csv entry
func newTransactionFromCSV(entry []string) (*ingTransaction, error) {
	bT, err := time.Parse("02.01.2006", entry[date])
	if err != nil {
		return nil, fmt.Errorf("could not parse date: %w", err)
	}

	vT, err := time.Parse("02.01.2006", entry[valueDate])
	if err != nil {
		return nil, fmt.Errorf("could not parse valueDate: %w", err)
	}

	sInt, err := moneyStringToInt(entry[saldo])
	if err != nil {
		return nil, fmt.Errorf("could not parse saldo to int: %w", err)
	}
	sMoney := money.New(int64(sInt), entry[sCurrency])

	bInt, err := moneyStringToInt(entry[amount])
	if err != nil {
		return nil, fmt.Errorf("could not parse amount to int: %w", err)
	}
	bMoney := money.New(int64(bInt), entry[aCurrency])

	return &ingTransaction{
		date:            bT,
		valueDate:       vT,
		client:          entry[client],
		transactionType: entry[transactionType],
		usage:           entry[usageLine],
		saldo:           sMoney,
		amount:          bMoney,
	}, nil
}

//createSalesLine creates :61: line for MT940 from transaction
func (t *ingTransaction) createSalesLine(writer io.Writer) error {
	// :61:_YYMMDD_MMDD_CD_00,00NTRFNONREF
	_, err := writer.Write(
		[]byte(fmt.Sprintf(":61:%s%s%s%sNTRFNONREF\r\n",
			t.date.Format("060102"),
			t.valueDate.Format("0102"),
			isCreditOrDebit(t.Amount()),
			swiftMoneyFormatter.Format(t.Amount().Absolute().Amount()),
		)),
	)

	if err != nil {
		return fmt.Errorf("could not create sales line: %w", err)
	}
	return nil
}

// createMultipurposeLine creates :86: line for MT940 from transaction
func (t *ingTransaction) createMultipurposeLine(writer io.Writer) error {

	gvcCode, ok := GVCCodes[t.transactionType]
	if !ok {
		return fmt.Errorf("could not find gvc code for text: %s", t.transactionType)
	}

	ag := "?32" + t.client
	if t.client == "" {
		ag = ""
	}

	u, err := convertUsageToFields(t.usage)
	if err != nil {
		return fmt.Errorf("could not convert usage line: %w", err)
	}

	lineStr := fmt.Sprintf(":86:%s?00%s%s%s\r\n", gvcCode, umlautsReplacer.Replace(t.transactionType), u, ag)
	if len(lineStr) > 390 {
		return fmt.Errorf("mulitpurpose line is too long")
	}

	//:86:999?00BuchungsText?20...?29Verwendungszweck?32Auftraggeber
	_, err = writer.Write([]byte(lineStr))
	if err != nil {
		return fmt.Errorf("could not create multipurpose line: %w", err)
	}

	return nil
}

// ConvertToMT940 converts transaction into MT940 format
func (t *ingTransaction) ConvertToMT940(writer io.Writer) error {
	err := t.createSalesLine(writer)
	if err != nil {
		return fmt.Errorf("could not convert ingTransaction to mt940: %w", err)
	}
	err = t.createMultipurposeLine(writer)
	if err != nil {
		return fmt.Errorf("could not convert ingTransaction to mt940: %w", err)
	}
	return nil
}

// getAccountNumber returns blz and accountNumber from meta tags of the ING csv
func getAccountNumber(meta []string) (string, string) {
	// get iban line and split it, iban is in the second row
	iban := strings.Split(meta[2], ";")[1]
	// replace all whitespaces
	iban = strings.ReplaceAll(iban, " ", "")
	// blz begins in position 4 and has 8 chars
	// accountNumber begins in position 12 and has 10 chars (until the end of iban)
	return iban[4:12], strings.TrimSpace(iban[12:])
}

// cleanUpTransactions removes the first line of the csv data, and reverses the order of the rest,
// ING displays all transactions in ascending order, we need descending for mt940
func cleanUpTransactions(ts [][]string) [][]string {
	// remove first entry, thats the header
	ts = ts[1:]

	// reverse transactions
	for i := 0; i < len(ts)/2; i++ {
		ts[i], ts[len(ts)-1-i] = ts[len(ts)-1-i], ts[i]
	}
	return ts
}

// extractMetaFields removes and returns the first 15 lines from the csv content,
// that are in case of the ing-Diba meta fields that are no transactions and only infos about the sheet
func extractMetaFields(b *bufio.Reader) ([]string, error) {
	var meta = make([]string, 0, 14)
	for i := 0; i < 14; i++ {
		line, err := b.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("could not read line %d: %w", i, err)
		}
		if line != "\n" {
			meta = append(meta, line)
		}
	}
	return meta, nil
}
