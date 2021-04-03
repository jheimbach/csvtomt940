package n26

import (
	"fmt"
	"io"
	"time"

	"github.com/JHeimbach/csvtomt940/converter"
	"github.com/Rhymond/go-money"
)

// column mapping of n26 csv file
const (
	date int = iota
	payee
	accountNumber
	transactionType
	reference
	category
	amount
	amountForeign
	foreignCurrency
	exchangeRate
)

var gvcCodes = map[string]string{
	"Income":                    "051",
	"Gutschrift":                "051",
	"Outgoing Transfer":         "020",
	"Überweisung":               "020",
	"MasterCard Payment Credit": "051", // incoming payments to credit card
	"MasterCard Zahlung Credit": "051",
	"MasterCard Payment Debit":  "004", // outgoing payments to credit card
	"MasterCard Zahlung Debit":  "004",
}

type n26Transaction struct {
	date            time.Time
	payee           string
	transactionType string
	category        string
	reference       string
	saldo           *money.Money
	amount          *money.Money
}

func (n *n26Transaction) ConvertToMT940(writer io.Writer) error {
	panic("implement me")
}

func (n *n26Transaction) Saldo() *money.Money {
	return n.saldo
}

func (n *n26Transaction) Amount() *money.Money {
	return n.amount
}

func (n *n26Transaction) Date() time.Time {
	return n.date
}

func newTransactionFromCsv(entry []string, startSaldo *money.Money) (*n26Transaction, *money.Money, error) {

	tDate, err := time.Parse("2006-01-02", entry[date])
	if err != nil {
		return nil, nil, fmt.Errorf("could not parse date from %s: %w", entry[date], err)
	}

	tAmount, err := converter.MoneyStringToInt(entry[amount])
	if err != nil {
		return nil, nil, fmt.Errorf("could not parse amount to int: %w", err)
	}
	tAmountMoney := money.New(int64(tAmount), "EUR")

	tType := entry[transactionType]
	if isMastercardPayment(tType) {
		if converter.IsDebit(tAmountMoney) {
			tType = tType + " Debit"
		} else {
			tType = tType + " Credit"
		}
	}

	payeeText := entry[payee]
	if len(payeeText) >= 54 {
		payeeText = converter.SplitStringInParts(payeeText, 54, false)[0]
	}

	saldo, err := tAmountMoney.Add(startSaldo)
	if err != nil {
		return nil, nil, fmt.Errorf("could not add startsaldo to amount: %w", err)
	}

	transaction := &n26Transaction{
		date:            tDate,
		payee:           payeeText,
		transactionType: tType,
		category:        entry[category],
		reference:       entry[reference],
		saldo:           saldo,
		amount:          tAmountMoney,
	}

	return transaction, saldo, nil
}

func isMastercardPayment(transactionType string) bool {
	if transactionType == "MasterCard Payment" || transactionType == "MasterCard Zahlung" {
		return true
	}
	return false
}
