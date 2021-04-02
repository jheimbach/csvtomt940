package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/JHeimbach/csvtomt940/banks/ing"
	"github.com/JHeimbach/csvtomt940/mt940"
	"golang.org/x/text/encoding/charmap"
)

func usage(programName string) string {
	return fmt.Sprintf("USAGE:\n\t %s <transactions.csv>", programName)
}

func main() {
	var ingHasCategory = flag.Bool("ing-has-category", true, "Set to false when ing csv has no category column")
	var bankType = flag.String("bank-type", "ing", "Which converter should be used (available options: ing, n26")
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

	bank := getBank(*bankType, *ingHasCategory)
	if bank == nil {
		log.Fatalf("banktype %s not found", *bankType)
	}
	bank.ParseCsv(b)

	// create sta file
	staFileName := strings.ReplaceAll(csvFileName, ".csv", ".sta")
	staFile, err := os.Create(staFileName)
	if err != nil {
		log.Fatalf("could not create file: %s: %v ", staFileName, err)
	}

	err = bank.ConvertToMT940(staFile)
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

func getBank(bankType string, ingHasCategory bool) mt940.Bank {
	switch bankType {
	case "ing":
		{
			return &ing.Ing{
				HasCategory: ingHasCategory,
			}
		}
	}
	return nil
}
