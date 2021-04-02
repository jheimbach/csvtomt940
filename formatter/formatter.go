package formatter

import "github.com/Rhymond/go-money"

// swiftMoneyFormatter formats money values according the specification for amount values in MT940
var swiftMoneyFormatter = money.NewFormatter(2, ",", "", "", "1")

func ConvertMoneyToString(m *money.Money) string {
	return swiftMoneyFormatter.Format(m.Amount())
}
