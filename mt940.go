package main

import (
	"fmt"
	"io"

	"github.com/Rhymond/go-money"
)

// MT940Converter converts csv transactions into the MT940 format
type MT940Converter interface {
	ConvertToMT940(writer io.Writer) error
}

// swiftTransactions creates a MT940 statement from given transactions
// accountNumber and bankNumber are required for the accountLine (:25:)
type swiftTransactions struct {
	accountNumber string
	bankNumber    string
	transactions  []Transaction
}

// swiftMoneyFormatter formats money values according the specification for amount values in MT940
var swiftMoneyFormatter = money.NewFormatter(2, ",", "", "", "1")

// createHeaderLine writes a headerline to the writer, it is static and returns always :20:CSVTOMT940
func (s *swiftTransactions) createHeaderLine(writer io.Writer) error {
	_, err := writer.Write([]byte(":20:CSVTOMT940\r\n"))

	if err != nil {
		return fmt.Errorf("could not create headerline: %w", err)
	}
	return nil
}

// createAccountLine creates account line :25: with bankNumber and accountNumber
func (s *swiftTransactions) createAccountLine(writer io.Writer) error {
	if s.bankNumber == "" {
		return fmt.Errorf("could not create account line with empty bankNumber")
	}
	if s.accountNumber == "" {
		return fmt.Errorf("could not create account line with empty accountNumber")
	}

	_, err := writer.Write([]byte(fmt.Sprintf(":25:%s/%s\r\n", s.bankNumber, s.accountNumber)))

	if err != nil {
		return fmt.Errorf("could not create account line: %w", err)
	}
	return nil
}

// createStatementLine writes the statementline :28:0 to the writer, it is static and does not change
func (s *swiftTransactions) createStatementLine(writer io.Writer) error {
	_, err := writer.Write([]byte(":28C:0\r\n"))

	if err != nil {
		return fmt.Errorf("could not create statement line: %w", err)
	}
	return nil
}

// createStartSaldoLine creates the start saldo line :60F: with help of the first transaction
func (s *swiftTransactions) createStartSaldoLine(writer io.Writer) error {
	if len(s.transactions) <= 0 {
		return fmt.Errorf("no transactions found, could not create start saldo line")
	}
	// get the first ingTransaction to calculate start saldo
	fTransaction := s.transactions[0]

	// subtract the amount from saldo to get the startSaldo
	startSaldo, err := fTransaction.Saldo().Subtract(fTransaction.Amount())
	if err != nil {
		return fmt.Errorf("could not calculate beginsaldo: %w", err)
	}

	// write line
	_, err = writer.Write([]byte(fmt.Sprintf(":60F:%s%s%s%s\r\n", isCreditOrDebit(startSaldo), fTransaction.Date().Format("060102"), startSaldo.Currency().Code, swiftMoneyFormatter.Format(startSaldo.Absolute().Amount()))))
	if err != nil {
		return fmt.Errorf("could not create begin startSaldo line: %w", err)
	}
	return nil
}

// createEndSaldoLine creates end saldo line :62F: with help of the last transaction
func (s *swiftTransactions) createEndSaldoLine(writer io.Writer) error {
	if len(s.transactions) <= 0 {
		return fmt.Errorf("no transactions found, could not create end saldo line")
	}
	lTransaction := s.transactions[len(s.transactions)-1]

	endSaldo := lTransaction.Saldo()

	_, err := writer.Write([]byte(fmt.Sprintf(":62F:%s%s%s%s", isCreditOrDebit(endSaldo), lTransaction.Date().Format("060102"), endSaldo.Currency().Code, swiftMoneyFormatter.Format(endSaldo.Absolute().Amount()))))

	if err != nil {
		return fmt.Errorf("could not create end saldo line: %w", err)
	}
	return nil
}

// ConvertToMT940 calls all line creation functions and writes a complete MT940 statement to the given writer
func (s *swiftTransactions) ConvertToMT940(w io.Writer) error {
	err := s.createHeaderLine(w)
	if err != nil {
		return err
	}
	err = s.createAccountLine(w)
	if err != nil {
		return err
	}
	err = s.createStatementLine(w)
	if err != nil {
		return err
	}
	err = s.createStartSaldoLine(w)
	if err != nil {
		return err
	}

	for i, t := range s.transactions {
		err = t.ConvertToMT940(w)
		if err != nil {
			return fmt.Errorf("could not convert transaction in line %d: %w", i, err)
		}
	}

	err = s.createEndSaldoLine(w)
	if err != nil {
		return err
	}

	_, err = w.Write([]byte("\r\n"))
	if err != nil {
		return fmt.Errorf("could not write last empty line: %w", err)
	}
	return nil
}
