package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/text/encoding/charmap"
)

func usage(programName string) string {
	return fmt.Sprintf("USAGE: %s\n <transactions.csv>", programName)
}

func main() {
	// if no file is given, return usage message
	if len(os.Args) < 2 {
		log.Fatalf(usage(os.Args[0]))
	}
	// get csv filename from arguments and open file
	csvFileName := os.Args[1]
	csvFile, err := os.Open(csvFileName)
	if err != nil {
		log.Printf("Could not open file %s", csvFileName)
	}
	defer csvFile.Close()

	// convert to utf8 because ing-diba encodes in ISO8859-1
	b := bufio.NewReader(charmap.ISO8859_1.NewDecoder().Reader(csvFile))

	// extract the first 14 lines from the reader, thats the meta infos
	meta, err := extractMetaFields(b)
	if err != nil {
		log.Fatalf("could not read meta fields: %v", err)
	}

	// extract banknumber and accountnumber from meta fields
	bankNumber, accountNumber := getAccountNumber(meta)

	sTransactions := &swiftTransactions{
		accountNumber: accountNumber,
		bankNumber:    bankNumber,
	}

	// read rest of the file as csv
	cr := csv.NewReader(b)
	cr.Comma = ';'

	transactions, err := cr.ReadAll()
	if err != nil {
		log.Fatalf("could not read transactions from csv %v", err)
	}
	// remove first line and reverse the order
	transactions = cleanUpTransactions(transactions)

	// create transaction structs
	var ta = make([]*transaction, 0, len(transactions))
	for i, t := range transactions {
		ts, err := newTransactionFromCSV(t)
		if err != nil {
			log.Fatalf("could not convert entry to struct in line %d: %v", i, err)
		}
		ta = append(ta, ts)
	}
	sTransactions.transactions = ta

	staFileName := strings.ReplaceAll(csvFileName, ".csv", ".sta")
	staFile, err := os.Create(staFileName)
	if err != nil {
		log.Fatalf("could not create file: %s: %v ", staFileName, err)
	}

	err = sTransactions.convertToMt940(staFile)
	if err != nil {
		log.Fatalf("could not convert to MT940: %v", err)
	}
	err = staFile.Close()
	if err != nil {
		log.Fatalf("could close file: %v", err)
	}
	log.Println("done")
}

func extractMetaFields(b *bufio.Reader) ([]string, error) {
	var meta = make([]string, 0, 14)
	for i := 0; i < 14; i++ {
		line, err := b.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("could not read line %d: %w", i, err)
		}
		if line != "\n" {
			meta = append(meta, line)
		}
	}
	return meta, nil
}

// cleanUpTransactions removes the first line of the csv data, and reverses the order of the rest,
// ING displays all transactions in ascending order, we need descending for mt940
func cleanUpTransactions(ts [][]string) [][]string {
	// remove first entry, thats the header
	ts = ts[1:]

	// reverse transactions
	for i := 0; i < len(ts)/2; i++ {
		ts[i], ts[len(ts)-1-i] = ts[len(ts)-1-i], ts[i]
	}
	return ts
}

// getAccountNumber returns blz and accountNumber from meta tags of the ING csv
func getAccountNumber(meta []string) (string, string) {
	// get iban line and split it, iban is in the second row
	iban := strings.Split(meta[2], ";")[1]
	// replace all whitespaces
	iban = strings.ReplaceAll(iban, " ", "")
	// blz begins in position 4 and has 8 chars
	// accountNumber begins in position 12 and has 10 chars (until the end of iban)
	return iban[4:12], strings.TrimSpace(iban[12:])
}
