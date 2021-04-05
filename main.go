package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/JHeimbach/csvtomt940/banks/ing"
	"github.com/JHeimbach/csvtomt940/banks/n26"
	"github.com/JHeimbach/csvtomt940/mt940"
)

func usage(programName string) string {
	return fmt.Sprintf("USAGE:\n\t %s <transactions.csv>", programName)
}

func main() {
	var ingHasCategory = flag.Bool("ing-has-category", true, "Set to false when ing csv has no category column")
	var bankType = flag.String("bank-type", "ing", "Which converter should be used (available options: ing, n26")
	var n26Iban = flag.String("n26-iban", "", "N26 does not save iban in csv export, you have to provide it yourself")

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

	bank, err := getBank(*bankType, *ingHasCategory, *n26Iban)
	if err != nil {
		log.Fatal(err)
	}
	bankInfos := bank.ParseCsv(csvFile)

	// create sta file
	staFileName := strings.ReplaceAll(csvFileName, ".csv", ".sta")
	staFile, err := os.Create(staFileName)
	if err != nil {
		log.Fatalf("could not create file: %s: %v ", staFileName, err)
	}

	err = bankInfos.ConvertToMT940(staFile)
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

func getBank(bankType string, ingHasCategory bool, iban string) (mt940.Bank, error) {
	switch bankType {
	case "ing":
		{
			return &ing.Ing{
				HasCategory: ingHasCategory,
			}, nil
		}
	case "n26":
		{
			if iban == "" {
				return nil, errors.New("parser for N26 need iban provided")
			}
			return &n26.N26{
				Iban: iban,
			}, nil
		}
	}
	return nil, fmt.Errorf("bank \"%s\" not supported", bankType)
}
