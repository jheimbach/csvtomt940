package main

import (
	"fmt"
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
		BuchungsText:     entry[Buchung],
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
	return strconv.Atoi(m)
}
