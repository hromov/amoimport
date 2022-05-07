package amoimport

import (
	"gorm.io/gorm"
)

type ImportService struct {
	DB *gorm.DB
}

func Import(db *gorm.DB, leads_path string, contacts_path string, n int) error {

	is := &ImportService{DB: db}

	if err := is.Push_Misc(leads_path, n); err != nil {
		return err
	}

	if err := is.Push_Contacts(contacts_path, n); err != nil {
		return err
	}

	if err := is.Push_Leads(leads_path, n); err != nil {
		return err
	}
	return nil
}
