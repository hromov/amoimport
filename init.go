package amoimport

import (
	"log"

	"gorm.io/gorm"
)

type ImportService struct {
	DB *gorm.DB
}

func Import(db *gorm.DB, folder string, n int) {

	is := &ImportService{DB: db}

	if err := is.Push_Misc(folder+"leads.csv", n); err != nil {
		log.Println(err)
	}

	if err := is.Push_Contacts(folder+"contacts.csv", n); err != nil {
		log.Println(err)
	}

	if err := is.Push_Leads(folder+"leads.csv", n); err != nil {
		log.Println(err)
	}
}
