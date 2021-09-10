# CSV to MT940
Converter for CSV to MT940

I needed a converter for my banking statements from an ing account. 
ing does not offer a mt940 download so i wrote my own converter. 
I do not guarantee it will work with your statements! 

If you have any problems please create an issue or a pull request

## Install
To install you have to have [go](https://golang.org/) installed on your machine.

Install it via go get:

```shell
go get github.com/JHeimbach/csvtomt940
```

## Usage
To run the converter run:
```shell
csvtomt940 SourceFile.csv
```

It will produce a .sta file with the same name as the given .csv file

## Flags
| name | default| usage|
|---|---|---|
|`-ing-has-category`| `true` | Use this if you want to use this converter with the old csv files from ing (that don't have a category entry), set this flag to false |
|`-bank-type`| `ing`| this program can convert the csv from ing and n26 bank|
|`-n26-iban`| `<none>` | n26 csv export does not include the account iban, but mt940 needs this, please provide your iban with this option
|`-n26-start-saldo`| `<none>` | n26 csv export does not include saldo infos, but mt940 needs this, please provide your startsaldo with this option in cents (e.g. 150,34€ is 15034)

## Example CSVs

### ING
:bulb: PLEASE NOTE: ING Csv files are expected to be in ISO-8859-1 Encoding, because that's what the csv export from ING is giving me.

#### Current Format
```csv
Umsatzanzeige;Datei erstellt am: 07.03.2021 13:18
;Letztes Update: aktuell

IBAN;DE32 5001 0517 1234 5678 95
Kontoname;Girokonto
Bank;ING
Kunde;Test Tester
Zeitraum;06.01.2020 - 09.01.2020
Saldo;1172,12;EUR

Sortierung;Datum absteigend

In der CSV-Datei finden Sie alle bereits gebuchten Umsätze. Die vorgemerkten Umsätze werden nicht aufgenommen, auch wenn sie in Ihrem Internetbanking angezeigt werden.

Buchung;Valuta;Auftraggeber/Empfänger;Buchungstext;Kategorie;Verwendungszweck;Saldo;Währung;Betrag;Währung
09.01.2020;09.01.2020;Yabox;Lastschrift;Shopping und Media;Reactive full-range local area network;1188,32;EUR;-1,62;EUR
06.01.2020;06.01.2020;Yabox;Gutschrift;Shopping und Media;Grass-roots systemic pricing structure;1172,12;EUR;16,20;EUR
```

#### Old Format without Categories
```csv
Umsatzanzeige;Datei erstellt am: 07.03.2021 13:18
;Letztes Update: aktuell

IBAN;DE32 5001 0517 1234 5678 95
Kontoname;Girokonto
Bank;ING
Kunde;Test Tester
Zeitraum;06.01.2020 - 09.01.2020
Saldo;1172,12;EUR

Sortierung;Datum absteigend

In der CSV-Datei finden Sie alle bereits gebuchten Umsätze. Die vorgemerkten Umsätze werden nicht aufgenommen, auch wenn sie in Ihrem Internetbanking angezeigt werden.

Buchung;Valuta;Auftraggeber/Empfänger;Buchungstext;Verwendungszweck;Saldo;Währung;Betrag;Währung
09.01.2020;09.01.2020;Yabox;Lastschrift;Reactive full-range local area network;1188,32;EUR;-1,62;EUR
06.01.2020;06.01.2020;Yabox;Gutschrift;Grass-roots systemic pricing structure;1172,12;EUR;16,20;EUR
```

### N26
:bulb: PLEASE NOTE: N26 CSV files do not have account infos in them, but we need an account and bank number that we extract from the iban, please provide your iban via `-n26-iban` option
```csv
"Datum","Empfänger","Kontonummer","Transaktionstyp","Verwendungszweck","Kategorie","Betrag (EUR)","Betrag (Fremdwährung)","Fremdwährung","Wechselkurs"
"2021-02-08","Yabox","DE00111111110000000000","Gutschrift","Grass-roots systemic pricing structure","Medien & Elektronik","16.2","","",""
"2021-02-08","Yabox","DE00111111110000000000","Lastschrift","Grass-roots systemic pricing structure","Medien & Elektronik","-1.62","","",""
```

you can also provide an english csv, it doesn't matter because we operate on the column index, not on the column name.
```csv
"Date","Payee","Account number","Transaction type","Payment reference","Category","Amount (EUR)","Amount (Foreign Currency)","Type Foreign Currency","Exchange Rate"
"2021-02-08","Yabox","DE00111111110000000000","Income","Grass-roots systemic pricing structure","Medien & Elektronik","16.2","","",""
"2021-02-08","Yabox","","Outgoing Transfer","Grass-roots systemic pricing structure","Medien & Elektronik","-1.62","","",""
```