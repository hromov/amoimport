package main

import (
	"log"
	"os"

	"github.com/hromov/amoimport/amoimport"
	"github.com/hromov/jevelina/auth"
	"github.com/hromov/jevelina/cdb"
)

const leads = "_import/amocrm_export_leads_2022-04-20.csv"
const contacts = "_import/amocrm_export_contacts_2022-04-20.csv"
const rowsToImport = 3000

func main() {
	dsn, err := os.ReadFile("../jevelina/_keys/db_local")
	if err != nil {
		log.Fatal(err)
	}

	db, err := cdb.OpenAndInit(string(dsn))
	if err != nil {
		log.Fatalf("Cant open and init data base error: %s", err.Error())
	}

	if err = auth.InitUsers(db.DB); err != nil {
		log.Fatalf("Can't init users error: %s", err.Error())
	}

	if err := amoimport.Import(db.DB, leads, contacts, rowsToImport); err != nil {
		log.Fatalf("Can't import error: %s", err.Error())
	}
}
