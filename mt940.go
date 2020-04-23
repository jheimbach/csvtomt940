package main

import (
	"fmt"
	"io"

	"github.com/Rhymond/go-money"
)

type swiftTransactions struct {
	accountNumber string
	bankNumber    string
	beginSaldo    string
	transactions  []*transaction
}

func (s *swiftTransactions) createHeaderLine(writer io.Writer) error {
	_, err := writer.Write([]byte(":20:CSVTOMT940\r\n"))

	if err != nil {
		return fmt.Errorf("could not create headerline: %w", err)
	}
	return nil
}

func (s *swiftTransactions) createAccountLine(writer io.Writer) error {
	_, err := writer.Write([]byte(fmt.Sprintf(":25:%s/%s\r\n", s.bankNumber, s.accountNumber)))

	if err != nil {
		return fmt.Errorf("could not create account line: %w", err)
	}
	return nil
}

func (s *swiftTransactions) createStatementLine(writer io.Writer) error {
	_, err := writer.Write([]byte(":28:0\r\n"))

	if err != nil {
		return fmt.Errorf("could not create statement line: %w", err)
	}
	return nil
}

func (s *swiftTransactions) createBeginSaldoLine(writer io.Writer) error {
	// get the first transaction to calculate start saldo
	fTransaction := s.transactions[0]

	// subtract the amount from saldo to get the startSaldo
	startSaldo, err := fTransaction.Saldo.Subtract(fTransaction.Betrag)
	if err != nil {
		return fmt.Errorf("could not calculate beginsaldo: %w", err)
	}

	// new money formater for this line
	mFormat := money.NewFormatter(2, ",", "", "", "1")

	//determine if value is credit or debit
	credtDebit := "C"
	if startSaldo.IsNegative() {
		credtDebit = "D"
	}

	// write line
	_, err = writer.Write([]byte(fmt.Sprintf(":60F:%s%s%s%s\r\n", credtDebit, fTransaction.Buchung.Format("060102"), startSaldo.Currency().Code, mFormat.Format(startSaldo.Amount()))))
	if err != nil {
		return fmt.Errorf("could not create begin startSaldo line: %w", err)
	}
	return nil
}
