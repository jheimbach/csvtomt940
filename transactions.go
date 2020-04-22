package main

import (
	"fmt"
	"time"
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
	Saldo            string
	SWaehrung        string
	Betrag           string
	BWaehrung        string
}

func newTransactionFromCSV(entry []string) (*transaction, error) {
	bT, err := time.Parse("02.01.2006", entry[Buchung])
	if err != nil {
		return nil, fmt.Errorf("could not parse Buchung: %w", err)
	}

	vT, err := time.Parse("02.01.2006", entry[Valuta])
	if err != nil {
		return nil, fmt.Errorf("could not parse Valuta: %w", err)
	}

	return &transaction{
		Buchung:          bT,
		Valuta:           vT,
		Auftraggeber:     entry[Auftraggeber],
		BuchungsText:     entry[Buchungstext],
		Verwendungszweck: entry[Verwendungszweck],
		Saldo:            entry[Saldo],
		SWaehrung:        entry[SWaehrung],
		Betrag:           entry[Betrag],
		BWaehrung:        entry[BWaehrung],
	}, nil
}
