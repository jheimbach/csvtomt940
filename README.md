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
go get github.com/jheimbach/csvtomt940
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