package main

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/Rhymond/go-money"
)

type ValuePosition int

const (
	Buchung ValuePosition = iota
	Valuta
	Auftraggeber
	Buchungstext
	Verwendungszweck
	Saldo
	SWaehrung
	Betrag
	BWaehrung
)

type transaction struct {
	Buchung          time.Time
	Valuta           time.Time
	Auftraggeber     string
	BuchungsText     string
	Verwendungszweck string
	Saldo            *money.Money
	Betrag           *money.Money
}

func (t *transaction) createSalesLine(writer io.Writer) error {
	// :60:_YYMMDD_MMDD_CD_00,00NTRFNONREF
	_, err := writer.Write([]byte(fmt.Sprintf(":61:%s%s%s%sNTRFNONREF\r\n", t.Buchung.Format("060102"), t.Valuta.Format("0102"), isCreditOrDebit(t.Betrag), swiftMoneyFormatter.Format(t.Betrag.Absolute().Amount()))))

	if err != nil {
		return fmt.Errorf("could not create sales line: %w", err)
	}
	return nil
}

func (t *transaction) createMultipurposeField(writer io.Writer) error {

	gvcCode, ok := GVCCodes[t.BuchungsText]
	if !ok {
		return fmt.Errorf("could not find gvc code for text: %s", t.BuchungsText)
	}

	ag := "?32" + t.Auftraggeber
	if t.Auftraggeber == "" {
		ag = ""
	}
	//:86:999?00BuchungsText?20...?29Verwendungszweck?32Auftraggeber
	_, err := writer.Write([]byte(fmt.Sprintf(":86:%s?00%s%s%s\r\n", gvcCode, t.BuchungsText, convertUsageToFields(t.Verwendungszweck), ag)))

	if err != nil {
		return fmt.Errorf("could not create multipurpose line: %w", err)
	}
	return nil
}

func (t *transaction) convertToMt940(writer io.Writer) error {
	err := t.createSalesLine(writer)
	if err != nil {
		return fmt.Errorf("could not convert transaction to mt940: %w", err)
	}
	err = t.createMultipurposeField(writer)
	if err != nil {
		return fmt.Errorf("could not convert transaction to mt940: %w", err)
	}
	return nil
}

func convertUsageToFields(usage string) string {
	usageWithControl := fmt.Sprintf("SVWZ+%s", usage)
	parts := splitStringInParts(usageWithControl, 27)
	if usage == "" {
		parts = []string{}
	}

	result := ""
	startControl := 20
	for _, part := range parts {
		result += fmt.Sprintf("?%d%s", startControl, strings.TrimSpace(part))
		startControl++
	}
	result += fmt.Sprintf("?%d%s", startControl, "KREF+NONREF")
	return result
}

func newTransactionFromCSV(entry []string) (*transaction, error) {
	bT, vT, err := parseTimeValues(entry)
	if err != nil {
		return nil, err
	}

	sMoney, bMoney, err := parseMoneyValues(entry)
	if err != nil {
		return nil, err
	}

	return &transaction{
		Buchung:          bT,
		Valuta:           vT,
		Auftraggeber:     entry[Auftraggeber],
		BuchungsText:     entry[Buchungstext],
		Verwendungszweck: entry[Verwendungszweck],
		Saldo:            sMoney,
		Betrag:           bMoney,
	}, nil
}

func parseMoneyValues(entry []string) (*money.Money, *money.Money, error) {
	sInt, err := moneyStringToInt(entry[Saldo])
	if err != nil {
		return nil, nil, fmt.Errorf("could not parse saldo to int: %w", err)
	}
	sMoney := money.New(int64(sInt), entry[SWaehrung])

	bInt, err := moneyStringToInt(entry[Betrag])
	if err != nil {
		return nil, nil, fmt.Errorf("could not parse betrag to int: %w", err)
	}
	bMoney := money.New(int64(bInt), entry[BWaehrung])
	return sMoney, bMoney, nil
}

func parseTimeValues(entry []string) (time.Time, time.Time, error) {
	bT, err := time.Parse("02.01.2006", entry[Buchung])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("could not parse Buchung: %w", err)
	}

	vT, err := time.Parse("02.01.2006", entry[Valuta])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("could not parse Valuta: %w", err)
	}
	return bT, vT, nil
}

func moneyStringToInt(m string) (int, error) {
	m = strings.ReplaceAll(m, ",", "")
	m = strings.ReplaceAll(m, ".", "")
	return strconv.Atoi(m)
}

// splitStringInParts cuts the string in pieces each l chars long
func splitStringInParts(s string, l int) []string {
	parts := make([]string, 0, int(math.Ceil(float64(len(s))/float64(l))))

	part := make([]rune, 0, l)
	i := 0
	for _, c := range s {
		part = append(part, c)
		if (i+1)%l == 0 {
			if c == ' ' {
				part = part[:len(part)-1]
				continue
			}
			parts = append(parts, string(part))
			part = make([]rune, 0, l)
		}
		i++
	}
	if len(part) > 0 {
		parts = append(parts, string(part))

	}

	return parts
}
