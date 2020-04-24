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

var swiftMoneyFormatter = money.NewFormatter(2, ",", "", "", "1")

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
	startSaldo, err := fTransaction.Saldo.Subtract(fTransaction.Amount)
	if err != nil {
		return fmt.Errorf("could not calculate beginsaldo: %w", err)
	}

	// write line
	_, err = writer.Write([]byte(fmt.Sprintf(":60F:%s%s%s%s\r\n", isCreditOrDebit(startSaldo), fTransaction.Buchung.Format("060102"), startSaldo.Currency().Code, swiftMoneyFormatter.Format(startSaldo.Absolute().Amount()))))
	if err != nil {
		return fmt.Errorf("could not create begin startSaldo line: %w", err)
	}
	return nil
}

func (s *swiftTransactions) createEndSaldoLine(writer io.Writer) error {
	lTransaction := s.transactions[len(s.transactions)-1]

	endSaldo := lTransaction.Saldo

	_, err := writer.Write([]byte(fmt.Sprintf(":62F:%s%s%s%s", isCreditOrDebit(endSaldo), lTransaction.Buchung.Format("060102"), endSaldo.Currency().Code, swiftMoneyFormatter.Format(endSaldo.Absolute().Amount()))))

	if err != nil {
		return fmt.Errorf("could not create end saldo line: %w", err)
	}
	return nil
}

func (s *swiftTransactions) convertToMt940(w io.Writer) error {
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
	err = s.createBeginSaldoLine(w)
	if err != nil {
		return err
	}

	for i, t := range s.transactions {
		err = t.createSalesLine(w)
		if err != nil {
			return fmt.Errorf("could not convert transaction in line %d: %w", i, err)
		}
		err = t.createMultipurposeField(w)
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

func isCreditOrDebit(amount *money.Money) string {
	//determine if value is credit or debit
	credtDebit := "C"
	if amount.IsNegative() {
		credtDebit = "D"
	}
	return credtDebit
}
