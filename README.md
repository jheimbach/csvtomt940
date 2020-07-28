# CSV to MT940
Converter for CSV to MT940

I needed a converter for my banking statements from an ing account. 
ing does not offer a mt940 download so i wrote my own converter. 
I do not guarantee it will work with your statements! 

If you have any problems please create an issue or a pull request

## Install
To install you have to have [go](https://golang.org/) installed on your machine.

Install it via go get:

```shell script
go get github.com/jheimbach/csvtomt940
```

## Usage
To run the converter run:
```shell script
csvtomt940 SourceFile.csv
```

It will produce a .sta file with the same name as the given .csv file

## Flags
| name | usage|
|---|---|
|`-old-syntax`| Use this if you want to use this converter with old csv files that don't have a category entry |