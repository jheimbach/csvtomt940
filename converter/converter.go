package converter

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/Rhymond/go-money"
)

// UmlautsReplacer replaces all umlauts with the two letter equivalent
var umlautsReplacer = strings.NewReplacer("Ä", "AE", "Ö", "OE", "Ü", "UE", "ß", "ss", "ä", "ae", "ö", "oe", "ü", "ue")

// ConvertUmlauts replaces all umlauts with the two letter equivalent
func ConvertUmlauts(s string) string {
	return umlautsReplacer.Replace(s)
}

// MoneyStringToInt converts formatted number to int, it removes every , and . and tries to parse the remaining string to a number
func MoneyStringToInt(m string) (int, error) {
	if m == "" {
		return 0, nil
	}
	m = strings.ReplaceAll(m, ",", "")
	m = strings.ReplaceAll(m, ".", "")
	return strconv.Atoi(m)
}

// ConvertUsageToFields splits the usage line in strings of 27 chars and adds control chars from ?20... to ?29
// if usage line is longer than 8*27 chars, it returns an error
func ConvertUsageToFields(usage string) (string, error) {
	parts := SplitStringInParts(fmt.Sprintf("SVWZ+%s", usage), 27, true)
	if usage == "" {
		parts = []string{}
	}
	if len(parts) > 8 {
		return "", fmt.Errorf("usage line is too long")
	}

	result, startControl := JoinFieldsWithControl(parts, 20)

	result += fmt.Sprintf("?%d%s", startControl, "KREF+NONREF")
	return result, nil
}

// JoinFieldsWithControl adds control number to the beginning of the line
func JoinFieldsWithControl(parts []string, startControl int) (string, int) {
	result := ""
	for _, part := range parts {
		result += fmt.Sprintf("?%d%s", startControl, strings.TrimSpace(part))
		startControl++
	}
	return result, startControl
}

// SplitStringInParts cuts the string in pieces each l chars long
func SplitStringInParts(s string, l int, trimWhitespace bool) []string {
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

func IsDebit(amount *money.Money) bool {
	return amount.IsNegative()
}

// IsCreditOrDebit returns a C if amount is positive and a D if amount is negative
func IsCreditOrDebit(amount *money.Money) string {
	//determine if value is credit or debit
	credtDebit := "C"
	if IsDebit(amount) {
		credtDebit = "D"
	}
	return credtDebit
}
