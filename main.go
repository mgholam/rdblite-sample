package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/mgholam/rdblite/storagefile"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {

	os.Mkdir("data/", 0755)
	os.Mkdir("json/", 0755)
	sf, err := storagefile.Open("data/invoices.dat")
	if err != nil {
		panic(err)
	}
	defer sf.Close()

	count := 100_000
	if sf.Count() == 0 {
		generateInvoices(sf, count)
	}
	// save to sqlite
	if !fileExists("data/invoices.db") {
		saveToSqlite(sf)
	}

	if !fileExists("json/table1.json") || !fileExists("json/invoices.json") {
		generateJsonFiles(sf)
	}

	sqlitetest()
	rdbtest()
}

func sqlitetest() {
	log.Println("sqlite test")
	sq, _ := gorm.Open(sqlite.Open("data/invoices.db?cache=shared&_pragma=journal_mode(wal)"), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})

	start := time.Now()
	var inv InvoiceGORM
	sq.Find(&inv, 99_999)
	log.Println("sqlite find by id 99,999 ", time.Since(start))

	start = time.Now()
	var invs []InvoiceGORM
	sq.Where("customer_name LIKE ?", "%Tomas%").Find(&invs)
	log.Println("sqlite where ", time.Since(start))
	log.Println("count", len(invs))

	fmt.Println()
	sqlDB, _ := sq.DB()
	sqlDB.Close()
}

func generateJsonFiles(sf *storagefile.StorageFile) {
	log.Println("generate json files")
	start := time.Now()
	var table1 []Table1
	var invoices []InvoiceTable

	for i := range sf.Iterate() {
		var inv Invoice
		r := Table1{}
		iv := InvoiceTable{}
		// unmarshal invoice
		e := json.Unmarshal(i.Data, &inv)
		if e != nil {
			log.Fatalln(e)
		}

		// create row from invoice data
		r.CustomerName = inv.CustomerName
		r.ID = inv.ID
		r.ItemCount = len(inv.Items)
		table1 = append(table1, r)

		iv.Address = inv.Address
		iv.CustomerName = inv.CustomerName
		iv.ID = inv.ID
		invoices = append(invoices, iv)
	}
	b, _ := json.MarshalIndent(table1, "", "  ")
	os.WriteFile("json/table1.json", b, 0644)

	b, _ = json.MarshalIndent(invoices, "", "  ")
	os.WriteFile("json/invoices.json", b, 0644)

	log.Println("end:", time.Since(start))
	fmt.Println()
}

func saveToSqlite(sf *storagefile.StorageFile) {
	// TODO : save to sqlite

	log.Println("sqlite start")
	start := time.Now()
	sq, err := gorm.Open(sqlite.Open("data/invoices.db?cache=shared&_pragma=journal_mode(wal)"), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})
	if err != nil {
		panic("failed to connect database")
	}
	sq.AutoMigrate(&InvoiceGORM{})

	tx := sq.Begin()
	for i := range sf.Iterate() {
		var iv InvoiceGORM
		json.Unmarshal(i.Data, &iv)

		tx.Create(&iv)
	}
	tx.Commit()
	log.Println("sqlite end", time.Since(start))
	fmt.Println()

	sqlDB, _ := sq.DB()
	sqlDB.Close()
}

func rdbtest() {
	db := NewDB()
	defer db.Close()

	rows := db.Table1.Query(func(row *Table1) bool {
		return strings.Contains(row.CustomerName, "Tomas") && row.ItemCount < 5
	})
	log.Println("query rows count =", len(rows))
	fmt.Println()

	str := "Bob"
	log.Println("search for", str)
	rows = db.Table1.Search(str)
	log.Println("search rows count =", len(rows))
	fmt.Println()

	// db.Table1.Delete(99999)
	log.Println("id 99,999 =", db.Table1.FindByID(99_999))
	fmt.Println()

	str = "Fort"
	log.Println("search for", str)
	inv := db.Invoices.Search(str)
	log.Println("search invoices count =", len(inv))
	fmt.Println()

	printMemUsage()
}

func generateInvoices(sf *storagefile.StorageFile, count int) {
	log.Println("generate invoices start")
	start := time.Now()
	for i := 0; i < count; i++ {
		item := Invoice{
			Date:         gofakeit.Date(),
			CustomerName: gofakeit.Name(),
			Address:      gofakeit.Address().Address,
		}
		item.ID = i + 1
		item.Items = []LineItem{}
		cc := rand.Intn(10)
		for j := 0; j < cc; j++ {
			item.Items = append(item.Items, LineItem{
				Product: fmt.Sprintf("Product %d", j+1),
				Qty:     gofakeit.Number(1, 100),
				Price:   gofakeit.Float32Range(100, 3000),
			})
		}
		b, _ := json.MarshalIndent(item, "", "   ")

		sf.Save("invoice", b)
	}
	log.Println("end:", time.Since(start))
	fmt.Println()
}

// -----------------------------------------------------------------------------

func printMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v MB", byteToMegaByte(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MB", byteToMegaByte(m.TotalAlloc))
	fmt.Printf("\tSys = %v MB", byteToMegaByte(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
	fmt.Println()
}

func byteToMegaByte(b uint64) uint64 {
	return b / 1024 / 1024
}
