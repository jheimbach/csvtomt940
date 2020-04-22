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
	if len(os.Args) < 2 {
		log.Fatalf(usage(os.Args[0]))
	}
	csvFileName := os.Args[1]
	csvFile, err := os.Open(csvFileName)
	if err != nil {
		log.Printf("Could not open file %s", csvFileName)
	}
	defer csvFile.Close()

	b := bufio.NewReader(charmap.ISO8859_1.NewDecoder().Reader(csvFile))
	var meta = make([]string, 0, 14)
	for i := 0; i < 14; i++ {
		line, err := b.ReadString('\n')
		if err != nil {
			log.Fatalf("read file line error: %v", err)
		}
		if line != "\n" {
			meta = append(meta, line)
		}
	}
	blz, accountNumber := getAccountNumber(meta)
	fmt.Println(blz)
	fmt.Println(accountNumber)

	cr := csv.NewReader(b)
	cr.Comma = ';'

	transactions, err := cr.ReadAll()
	if err != nil {
		log.Fatalf("could not read transactions from csv %v", err)
	}
	fmt.Println(transactions[:5])
	transactions = cleanUpTransactions(transactions)

	var ta = make([]*transaction, 0, len(transactions))
	for i, t := range transactions {
		ts, err := newTransactionFromCSV(t)
		if err != nil {
			log.Fatalf("could not convert entry to struct in line %d: %v", i, err)
		}
		ta = append(ta, ts)
	}
	fmt.Printf("%+v", ta[0])
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
	return iban[4:12], iban[12:]
}
