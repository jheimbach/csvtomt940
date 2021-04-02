package ing

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/JHeimbach/csvtomt940/converter"
	"github.com/JHeimbach/csvtomt940/formatter"
	"github.com/Rhymond/go-money"
)

// column mapping of ing csv file
const (
	date int = iota
	valueDate
	client
	transactionType
	category
	usageLine
	saldo
	sCurrency
	amount
	aCurrency
)

// GVCCodes returns the GVC Code for the given transactionType, note this list is not complete, other values are possible
var GVCCodes = map[string]string{
	"Abschluss":                         "805",
	"Gutschrift aus Dauerauftrag":       "052",
	"Abbuchung":                         "004",
	"Lastschrift":                       "005",
	"Gutschrift":                        "051",
	"Gehalt/Rente":                      "053",
	"Überweisung":                       "020",
	"Entgelt":                           "808",
	"Retouren":                          "059",
	"Dauerauftrag / Terminueberweisung": "008",
}

// ingTransaction is the implementation of Transaction for ing-diba csv format
type ingTransaction struct {
	date            time.Time
	valueDate       time.Time
	client          string
	transactionType string
	category        string
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
func newTransactionFromCSV(entry []string, hasCategory bool) (*ingTransaction, error) {
	var offset = 0
	if !hasCategory {
		offset = -1
	}
	bT, err := time.Parse("02.01.2006", entry[date])
	if err != nil {
		return nil, fmt.Errorf("could not parse date: %w", err)
	}

	vT, err := time.Parse("02.01.2006", entry[valueDate])
	if err != nil {
		return nil, fmt.Errorf("could not parse valueDate: %w", err)
	}

	sInt, err := converter.MoneyStringToInt(entry[saldo+offset])
	if err != nil {
		return nil, fmt.Errorf("could not parse saldo to int: %w", err)
	}
	sMoney := money.New(int64(sInt), entry[sCurrency+offset])

	bInt, err := converter.MoneyStringToInt(entry[amount+offset])
	if err != nil {
		return nil, fmt.Errorf("could not parse amount to int: %w", err)
	}
	bMoney := money.New(int64(bInt), entry[aCurrency+offset])

	cText := entry[client]
	if len(cText) >= 54 {
		cText = converter.SplitStringInParts(cText, 54, false)[0]
	}
	transaction := &ingTransaction{
		date:            bT,
		valueDate:       vT,
		client:          cText,
		transactionType: entry[transactionType],
		usage:           entry[usageLine+offset],
		saldo:           sMoney,
		amount:          bMoney,
	}
	if hasCategory {
		transaction.category = entry[category]
	}

	return transaction, nil
}

//createSalesLine creates :61: line for MT940 from transaction
func (t *ingTransaction) createSalesLine(writer io.Writer) error {
	// :61:<ValueDate><Date><IsCreditOrDebit><Amount>
	// :61:_YYMMDD_MMDD_CD_00,00NTRFNONREF
	_, err := writer.Write(
		[]byte(fmt.Sprintf(":61:%s%s%s%sNTRFNONREF\r\n",
			t.valueDate.Format("060102"),
			t.date.Format("0102"),
			converter.IsCreditOrDebit(t.Amount()),
			formatter.ConvertMoneyToString(t.Amount().Absolute()),
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

	c, _ := converter.JoinFieldsWithControl(converter.SplitStringInParts(t.client, 27, true), 32)
	if t.client == "" {
		c = ""
	}

	u, err := converter.ConvertUsageToFields(t.usage)
	if err != nil {
		return fmt.Errorf("could not convert usage line: %w", err)
	}

	lineStr := fmt.Sprintf("%s?00%s%s%s", gvcCode, converter.ConvertUmlauts(t.transactionType), u, c)
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
