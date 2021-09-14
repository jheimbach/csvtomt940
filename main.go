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
	var n26StartSaldo = flag.Int64("n26-start-saldo", 0, "N26 does not save saldo infos in csv export, you have to provide the startsaldo yourself, in cents e.g. 10,45€ = 1045")

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

	bank, err := getBank(*bankType, *ingHasCategory, *n26Iban, *n26StartSaldo)
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

func getBank(bankType string, ingHasCategory bool, iban string, saldo int64) (mt940.Bank, error) {
	switch bankType {
	case "ing":
		{
			return ing.New(ingHasCategory), nil
		}
	case "n26":
		{
			if iban == "" {
				return nil, errors.New("parser for N26 needs iban provided")
			}
			if saldo == 0 {
				log.Println("WARNING: N26 has no Saldo in its transaction statements, do you mean to start with saldo = 0?")
			}
			return n26.New(iban, saldo), nil
		}
	}
	return nil, fmt.Errorf("bank \"%s\" not supported", bankType)
}
