package ing

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/JHeimbach/csvtomt940/mt940"
	"golang.org/x/text/encoding/charmap"
)

type Ing struct {
	HasCategory bool
	data        *mt940.BankData
	logger      *log.Logger
}

func New(hasCategory bool) *Ing {
	logger := log.New(os.Stdout, "[ING] ", log.Lmsgprefix)

	return &Ing{
		logger:      logger,
		HasCategory: hasCategory,
	}
}

func (i *Ing) ParseCsv(csvFile *os.File) *mt940.BankData {
	// convert to utf8 because ing-diba encodes in ISO8859-1
	b := bufio.NewReader(charmap.ISO8859_1.NewDecoder().Reader(csvFile))

	// extract the first 14 lines from the reader, thats the meta infos
	meta, err := extractMetaFields(b)
	if err != nil {
		i.logger.Fatalf("could not read meta fields: %v", err)
	}

	// extract banknumber and accountnumber from meta fields
	bankNumber, accountNumber, err := getAccountNumber(meta)

	if err != nil {
		i.logger.Fatalf("could not get account number: %v", err)
	}

	i.data = &mt940.BankData{
		AccountNumber: accountNumber,
		BankNumber:    bankNumber,
	}

	// read rest of the file as csv
	cr := csv.NewReader(b)
	cr.Comma = ';'

	transactions, err := cr.ReadAll()
	if err != nil {
		i.logger.Fatalf("could not read data from csv %v", err)
	}
	// remove first line and reverse the order
	transactions = cleanUpTransactions(transactions)

	// create ingTransaction structs
	var ta = make([]mt940.Transaction, 0, len(transactions))
	for j, t := range transactions {
		ts, err := newTransactionFromCSV(t, i.HasCategory)
		if err != nil {
			i.logger.Fatalf("could not convert entry to struct in line %d: %v", j, err)
		}
		ta = append(ta, ts)
	}

	i.data.Transactions = ta

	return i.data
}

// extractMetaFields removes and returns the first 14 lines from the csv content,
// that are in case of the ing-Diba meta fields that are no data and only infos about the sheet
func extractMetaFields(b *bufio.Reader) ([]string, error) {
	var metablocksize = 13
	var meta = make([]string, 0, metablocksize)

	for i := 0; i < metablocksize; i++ {
		line, err := b.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil, fmt.Errorf("file format incorrect, file has only %d lines, meta block is %d lines long", i, metablocksize)
			}

			return nil, fmt.Errorf("could not read line %d: %w", i, err)
		}
		if line != "\n" {
			meta = append(meta, line)
		}
	}
	return meta, nil
}

// getAccountNumber returns blz and accountNumber from meta tags of the ING csv
func getAccountNumber(meta []string) (string, string, error) {
	// get iban line and split it, iban is in the second row
	metafields := strings.Split(meta[1], ";")

	if len(metafields) < 2 {
		return "", "", fmt.Errorf("could not split meta field %s", metafields[0])
	}
	iban := metafields[1]
	// replace all whitespaces
	iban = strings.ReplaceAll(iban, " ", "")
	// blz begins in position 4 and has 8 chars
	// accountNumber begins in position 12 and has 10 chars (until the end of iban)
	return iban[4:12], strings.TrimSpace(iban[12:]), nil
}

// cleanUpTransactions removes the first line of the csv data, and reverses the order of the rest,
// ING displays all data in ascending order, we need descending for mt940
func cleanUpTransactions(ts [][]string) [][]string {
	// remove first entry, thats the header
	ts = ts[1:]

	// reverse data
	for i := 0; i < len(ts)/2; i++ {
		ts[i], ts[len(ts)-1-i] = ts[len(ts)-1-i], ts[i]
	}
	return ts
}
