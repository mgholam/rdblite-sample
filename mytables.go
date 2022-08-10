package main

import (
	"time"

	"github.com/mgholam/rdblite"
)

type Table1 struct {
	rdblite.BaseTable
	CustomerName string
	ItemCount    int
}

// Invoice : the complete struct
type Invoice struct {
	rdblite.BaseTable
	Date         time.Time
	CustomerName string
	Address      string
	Items        []LineItem
}

type LineItem struct {
	Product string
	Qty     int
	Price   float32
}

// InvoiceTable : invoice data as a table for querying
type InvoiceTable struct {
	rdblite.BaseTable
	CustomerName string
	Address      string
}

type InvoiceGORM struct {
	ID           int
	CustomerName string
	Address      string
}
