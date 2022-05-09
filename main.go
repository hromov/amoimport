package main

import (
	"log"

	"github.com/hromov/amoimport/amoimport"
	"github.com/hromov/jevelina/cdb"
)

const (
	leads        = "_import/amocrm_export_leads_2022-04-20.csv"
	contacts     = "_import/amocrm_export_contacts_2022-04-20.csv"
	rowsToImport = 0
	dsn          = "root:password@tcp(127.0.0.1:3306)/gorm_test?parseTime=True&charset=utf8mb4"
)

func main() {
	db, err := cdb.OpenAndInit(string(dsn))
	if err != nil {
		log.Fatalf("Cant open and init data base error: %s", err.Error())
	}

	if err := amoimport.Import(db.DB, leads, contacts, rowsToImport); err != nil {
		log.Fatalf("Can't import error: %s", err.Error())
	}
}
