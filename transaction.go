package main

import (
	"github.com/Rhymond/go-money"
	"time"
)

// Transaction is the interface for each transaction line, it should convert to a valid mt940 string with lines 61 and 86
// the methods saldo, amount and date are used for creating the start and end saldo lines
type Transaction interface {
	MT940Converter
	Saldo() *money.Money
	Amount() *money.Money
	Date() time.Time
}
