# CSV to MT940
Convert CSV Bank Statement exports from ing (ing.de) or n26 to mt940 format.

I needed a converter for my banking statements from an ing account. 
ing does not offer a mt940 download, so I wrote my own converter. 
I do not guarantee it will work with your statements! 

If you have any problems please create an issue or a pull request

## Use Docker image
```shell
docker run --rm -it -v "$(pwd):/app" jheimbach/csvtomt940:latest [flags] sourcefile.csv
```


Note: docker creates files with root permissions, if this is a problem, simply run docker as a different user with
(User ID 1000 is the default linux user, check the `id` command which id is used for your account)
```shell
docker run --rm -it -u 1000:1000 -v "$(pwd):/app" jheimbach/csvtomt940:latest [flags] sourcefile.csv
```

## Use Release Builds
### Install
1. Download the latest Build from [Release Page](https://github.com/JHeimbach/csvtomt940/releases/latest)
2. extract the archive
   1. Optional: Move csvtomt940 executable to location in $PATH (e.g. `/usr/local/bin`)
### Usage
```bash
/path/to/csvtomt940 [flags] sourcefile.csv
```

## Use from Source code

### Install
To install you have to have [go](https://golang.org/) installed on your machine.

Install it via go get:

```shell
go get github.com/JHeimbach/csvtomt940
```

### Usage
run the converter with:
```shell
csvtomt940 sourcefile.csv
```

If you want to run with flags, make sure to add them in front of the csv file e.g:

```shell
csvtomt940 -bank-type n26 -n26-iban DEXXXX --n26-start-saldo XXXX sourcefile.csv
```

It will produce a .sta file with the same name as the given .csv file

## Flags
| name                | default  | required                | usage                                                                                                                                                                                                                                |
|---------------------|----------|-------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `-ing-has-category` | `true`   | No                      | _[DEPRECATED] - use has-category instead_ <br/>Set to false when ing csv has no category columnUse this if you want to use this converter with the old csv files from ing (that don't have a category entry), set this flag to false |
| `-has-category`     | `true`   | No                      | Use this if you want to use this converter with csv files that include a category column                                                                                                                                             |
| `-bank-type`        | `ing`    | Yes                     | this program can convert the csv from ing and n26 bank                                                                                                                                                                               |
| `-n26-iban`         | `<none>` | if `bank-type` is `n26` | n26 csv export does not include the account iban, but mt940 needs this, please provide your iban with this option                                                                                                                    |
| `-n26-start-saldo`  | `<none>` | if `bank-type` is `n26` | n26 csv export does not include saldo infos, but mt940 needs this, please provide your startsaldo with this option in cents (e.g. 150,34€ is 15034)                                                                                  |

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