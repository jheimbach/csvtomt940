package mt940

import (
	"fmt"
	"io"
	"os"

	"github.com/JHeimbach/csvtomt940/converter"
	"github.com/JHeimbach/csvtomt940/formatter"
)

// Converter converts csv transactions into the MT940 format
type Converter interface {
	ConvertToMT940(writer io.Writer) error
}

type Bank interface {
	ParseCsv(csvFile *os.File) *BankData
}

// swiftTransactions creates a MT940 statement from given transactions
// accountNumber and bankNumber are required for the accountLine (:25:)
type BankData struct {
	AccountNumber string
	BankNumber    string
	Transactions  []Transaction
}

// createHeaderLine writes a headerline to the writer, it is static and returns always :20:CSVTOMT940
func (s *BankData) createHeaderLine(writer io.Writer) error {
	_, err := writer.Write([]byte(":20:CSVTOMT940\r\n"))

	if err != nil {
		return fmt.Errorf("could not create headerline: %w", err)
	}
	return nil
}

// createAccountLine creates account line :25: with BankNumber and AccountNumber
func (s *BankData) createAccountLine(writer io.Writer) error {
	if s.BankNumber == "" {
		return fmt.Errorf("could not create account line with empty bankNumber")
	}
	if s.AccountNumber == "" {
		return fmt.Errorf("could not create account line with empty accountNumber")
	}

	// :25:<BankNumber><AccountNumber>
	_, err := writer.Write(
		[]byte(
			fmt.Sprintf(
				":25:%s/%s\r\n",
				s.BankNumber,
				s.AccountNumber,
			),
		),
	)

	if err != nil {
		return fmt.Errorf("could not create account line: %w", err)
	}
	return nil
}

// createStatementLine writes the statementline :28:0 to the writer, it is static and does not change
func (s *BankData) createStatementLine(writer io.Writer) error {
	_, err := writer.Write([]byte(":28C:0\r\n"))

	if err != nil {
		return fmt.Errorf("could not create statement line: %w", err)
	}
	return nil
}

// createStartSaldoLine creates the start saldo line :60F: with help of the first transaction
func (s *BankData) createStartSaldoLine(writer io.Writer) error {
	if len(s.Transactions) <= 0 {
		return fmt.Errorf("no transactions found, could not create start saldo line")
	}
	// get the first ingTransaction to calculate start saldo
	fTransaction := s.Transactions[0]

	// subtract the amount from saldo to get the startSaldo
	startSaldo, err := fTransaction.Saldo().Subtract(fTransaction.Amount())
	if err != nil {
		return fmt.Errorf("could not calculate beginsaldo: %w", err)
	}

	// :60F:<DebitOrCredit><Date><Currency><Amount>
	_, err = writer.Write(
		[]byte(
			fmt.Sprintf(
				":60F:%s%s%s%s\r\n",
				converter.IsCreditOrDebit(startSaldo),
				fTransaction.Date().Format("060102"),
				startSaldo.Currency().Code,
				formatter.ConvertMoneyToString(startSaldo.Absolute()),
			),
		),
	)
	if err != nil {
		return fmt.Errorf("could not create begin startSaldo line: %w", err)
	}
	return nil
}

// createEndSaldoLine creates end saldo line :62F: with help of the last transaction
func (s *BankData) createEndSaldoLine(writer io.Writer) error {
	if len(s.Transactions) <= 0 {
		return fmt.Errorf("no transactions found, could not create end saldo line")
	}
	lTransaction := s.Transactions[len(s.Transactions)-1]

	endSaldo := lTransaction.Saldo()

	// :62F:<DebitOrCredit><Date><Currency><Amount>
	_, err := writer.Write(
		[]byte(
			fmt.Sprintf(
				":62F:%s%s%s%s",
				converter.IsCreditOrDebit(endSaldo),
				lTransaction.Date().Format("060102"),
				endSaldo.Currency().Code,
				formatter.ConvertMoneyToString(endSaldo.Absolute()),
			),
		),
	)

	if err != nil {
		return fmt.Errorf("could not create end saldo line: %w", err)
	}
	return nil
}

// ConvertToMT940 calls all line creation functions and writes a complete MT940 statement to the given writer
func (s *BankData) ConvertToMT940(w io.Writer) error {
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

	for i, t := range s.Transactions {
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
