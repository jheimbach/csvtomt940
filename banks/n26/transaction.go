package n26

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/JHeimbach/csvtomt940/converter"
	"github.com/JHeimbach/csvtomt940/formatter"
	"github.com/Rhymond/go-money"
)

// column mapping of n26 csv file
const (
	date int = iota
	valueDate
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
	"Credit Transfer":           "051",
	"Outgoing Transfer":         "020",
	"Überweisung":               "020",
	"Debit Transfer":            "020",
	"Presentment":               "020",
	"Lastschrift":               "005",
	"Direct Debit":              "005",
	"MasterCard Payment Credit": "051", // incoming payments to credit card
	"MasterCard Zahlung Credit": "051",
	"MasterCard Payment Debit":  "004", // outgoing payments to credit card
	"MasterCard Zahlung Debit":  "004",
	"N26 Empfehlung":            "051", // N26 Cashback
	"Reward":                    "051",
	"N26 Referral":              "051",
	"Fee":                       "808",
	"Presentment Refund":        "059",
}

type n26Transaction struct {
	date                  time.Time
	payee                 string
	transactionType       string
	transactionTypeLookup string
	category              string
	reference             string
	saldo                 *money.Money
	amount                *money.Money
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

func newTransactionFromCsv(entry []string, startSaldo *money.Money, hasCategory bool) (*n26Transaction, *money.Money, error) {
	var offset = 0
	if !hasCategory {
		offset = -1
	}

	tDate, err := time.Parse("2006-01-02", entry[date])
	if err != nil {
		return nil, nil, fmt.Errorf("could not parse date from %s: %w", entry[date], err)
	}

	tAmount, err := converter.MoneyStringToInt(getAmount(entry[amount+offset]))
	if err != nil {
		return nil, nil, fmt.Errorf("could not parse amount to int: %w", err)
	}
	tAmountMoney := money.New(int64(tAmount), "EUR")

	tType := entry[transactionType]
	ttLookup := tType
	if isMastercardPayment(tType) {
		if converter.IsDebit(tAmountMoney) {
			ttLookup = tType + " Debit"
		} else {
			ttLookup = tType + " Credit"
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
		date:                  tDate,
		payee:                 payeeText,
		transactionType:       tType,
		transactionTypeLookup: ttLookup,
		category:              entry[category],
		reference:             entry[reference],
		saldo:                 saldo,
		amount:                tAmountMoney,
	}

	return transaction, saldo, nil
}

func isMastercardPayment(transactionType string) bool {
	if transactionType == "MasterCard Payment" || transactionType == "MasterCard Zahlung" {
		return true
	}
	return false
}

func getAmount(entryAmount string) string {
	decimalPosition := strings.Index(entryAmount, ".")
	if len(entryAmount)-decimalPosition <= 2 {
		return entryAmount + "0"
	}
	return entryAmount
}

func (n *n26Transaction) ConvertToMT940(writer io.Writer) error {
	err := n.createSalesLine(writer)
	if err != nil {
		return err
	}

	err = n.createMultipurposeLine(writer)
	return err
}

func (n *n26Transaction) createSalesLine(writer io.Writer) error {
	// :61:<ValueDate><Date><IsCreditOrDebit><Amount>NTRFNONREF
	// :61:_YYMMDD_MMDD_C/D_00,00NTRFNONREF
	_, err := writer.Write(
		[]byte(fmt.Sprintf(":61:%s%s%s%sNTRFNONREF\r\n",
			n.date.Format("060102"),
			n.date.Format("0102"),
			converter.IsCreditOrDebit(n.Amount()),
			formatter.ConvertMoneyToString(n.Amount().Absolute()),
		)),
	)

	if err != nil {
		return fmt.Errorf("could not create sales line: %w", err)
	}
	return nil
}

// createMultipurposeLine creates :86: line for MT940 from transaction
func (n *n26Transaction) createMultipurposeLine(writer io.Writer) error {

	gvcCode, ok := gvcCodes[n.transactionTypeLookup]
	if !ok {
		return fmt.Errorf("could not find gvc code for text: %s", n.transactionTypeLookup)
	}

	c, _ := converter.JoinFieldsWithControl(converter.SplitStringInParts(n.payee, 27, true), 32)
	if n.payee == "" {
		c = ""
	}

	u, err := converter.ConvertUsageToFields(n.reference)
	if err != nil {
		return fmt.Errorf("could not convert reference line: %w", err)
	}

	lineStr := fmt.Sprintf("%s?00%s%s%s", gvcCode, converter.ConvertUmlauts(n.transactionType), u, c)
	if len(lineStr) > 390 {
		return fmt.Errorf("mulitpurpose line is too long")
	}
	lineParts := converter.SplitStringInParts(lineStr, 65, false)

	// :86:<GVCCode>?00<GVCText>?20..29<MEMO>?32<Payee>
	//:86:999?00BuchungsText?20...?29Verwendungszweck?32Auftraggeber
	_, err = writer.Write(
		[]byte(
			fmt.Sprintf(
				":86:%s\r\n",
				strings.Join(lineParts, "\r\n"),
			),
		),
	)
	if err != nil {
		return fmt.Errorf("could not create multipurpose line: %w", err)
	}

	return nil
}
