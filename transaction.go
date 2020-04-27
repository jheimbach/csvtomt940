package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/Rhymond/go-money"
)

// Transaction is the interface for each transaction line, it should convert to a valid mt940 string with lines 61 and 86
// the methods saldo, amount and date are used for creating the start and end saldo lines
type Transaction interface {
	MT940Converter
	Saldo() *money.Money
	Amount() *money.Money
	Date() time.Time
}

// umlautsReplacer replaces all umlauts with the two letter equivalent
var umlautsReplacer = strings.NewReplacer("Ä", "AE", "Ö", "OE", "Ü", "UE", "ß", "ss", "ä", "ae", "ö", "oe", "ü", "ue")

// converts formatted number to int, it removes every , and . and tries to parse the remaining string to a number
func moneyStringToInt(m string) (int, error) {
	if m == "" {
		return 0, nil
	}
	m = strings.ReplaceAll(m, ",", "")
	m = strings.ReplaceAll(m, ".", "")
	return strconv.Atoi(m)
}

// convertUsageToFields splits the usage line in strings of 27 chars and adds control chars from ?20... to ?29
// if usage line is longer than 8*27 chars, it returns an error
func convertUsageToFields(usage string) (string, error) {
	parts := splitStringInParts(fmt.Sprintf("SVWZ+%s", usage), 27, true)
	if usage == "" {
		parts = []string{}
	}
	if len(parts) > 8 {
		return "", fmt.Errorf("usage line is too long")
	}

	result := ""
	startControl := 20
	for _, part := range parts {
		result += fmt.Sprintf("?%d%s", startControl, strings.TrimSpace(part))
		startControl++
	}
	result += fmt.Sprintf("?%d%s", startControl, "KREF+NONREF")
	return result, nil
}

// splitStringInParts cuts the string in pieces each l chars long
func splitStringInParts(s string, l int, trimWhitespace bool) []string {
	parts := make([]string, 0, int(math.Ceil(float64(len(s))/float64(l))))

	part := make([]rune, 0, l)
	i := 0
	for _, c := range s {
		part = append(part, c)
		if (i+1)%l == 0 {
			if c == ' ' && trimWhitespace {
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

// isCreditOrDebit returns a C if amount is positive and a D if amount is negative
func isCreditOrDebit(amount *money.Money) string {
	//determine if value is credit or debit
	credtDebit := "C"
	if amount.IsNegative() {
		credtDebit = "D"
	}
	return credtDebit
}
