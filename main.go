package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/text/encoding/charmap"
)

func usage(programName string) string {
	return fmt.Sprintf("USAGE:\n\t %s <transactions.csv>", programName)
}

func main() {
	var oldSyntaxFlag = flag.Bool("old-syntax", false, "Use old CSV syntax, (without category column)")
	flag.Parse()
	// if no file is given, return usage message
	if len(os.Args) < 2 {
		log.Fatalf(usage(os.Args[0]))
	}
	// get csv filename from arguments and open file
	csvFileName := flag.Arg(0)
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

	// create ingTransaction structs
	var ta = make([]Transaction, 0, len(transactions))
	for i, t := range transactions {
		ts, err := newTransactionFromCSV(t, !(*oldSyntaxFlag))
		if err != nil {
			log.Fatalf("could not convert entry to struct in line %d: %v", i, err)
		}
		ta = append(ta, ts)
	}
	sTransactions.transactions = ta

	// create sta file
	staFileName := strings.ReplaceAll(csvFileName, ".csv", ".sta")
	staFile, err := os.Create(staFileName)
	if err != nil {
		log.Fatalf("could not create file: %s: %v ", staFileName, err)
	}

	// convert transactions to mt940 format
	err = sTransactions.ConvertToMT940(staFile)
	if err != nil {
		log.Fatalf("could not convert to MT940: %v", err)
	}
	// close the sta file
	err = staFile.Close()
	if err != nil {
		log.Fatalf("could close file: %v", err)
	}
	log.Println("done")
}
