package main

import (
	"fmt"
	"os"

	"github.com/mgholam/rdblite"
)

type DB struct {
	Table1   *rdblite.Table[Table1]
	Invoices *rdblite.Table[InvoiceTable]
}

func (d *DB) Close() {
	// save all tables
	d.Table1.Close()
	d.Invoices.Close()
}

func NewDB() *DB {

	db := DB{}

	db.Table1 = &rdblite.Table[Table1]{
		GobFilename: "data/table1.gob",
	}

	db.Invoices = &rdblite.Table[InvoiceTable]{
		GobFilename: "data/invoices.gob",
	}

	if fileExists(db.Table1.GobFilename) {
		db.Table1.LoadGob()
	} else {
		db.Table1.LoadJson("json/table1.json")
	}

	if fileExists(db.Invoices.GobFilename) {
		db.Invoices.LoadGob()
	} else {
		db.Invoices.LoadJson("json/invoices.json")
	}

	fmt.Println()

	return &db
}

func fileExists(fn string) bool {
	_, e := os.Stat(fn)
	return e == nil
}
