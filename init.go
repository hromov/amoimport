package amoimport

import (
	"os"

	"gorm.io/gorm"
)

const broken_dir = "broken"
const broken_leads = broken_dir + "/broken_leads.csv"
const broken_contacts = broken_dir + "/broken_contacts.csv"

type AmoService struct {
	DB            *gorm.DB
	sources       map[string]uint8
	users         map[string]uint64
	products      map[string]uint32
	manufacturers map[string]uint16
	steps         map[string]uint8
	tags          map[string]uint8
	//key = hash, val = id
	contacts map[string]uint64
	misc     map[string]bool
}

func Import(db *gorm.DB, leads_path string, contacts_path string, n int) error {
	err := os.MkdirAll(broken_dir, 0755)
	if err != nil {
		return err
	}

	amo := &AmoService{
		DB:            db,
		sources:       make(map[string]uint8),
		users:         make(map[string]uint64),
		products:      make(map[string]uint32),
		manufacturers: make(map[string]uint16),
		steps:         make(map[string]uint8),
		tags:          make(map[string]uint8),
		contacts:      make(map[string]uint64),
		misc:          make(map[string]bool),
	}

	if err := amo.Push_Misc(leads_path, n); err != nil {
		return err
	}

	if err := amo.LoadMiscsToMaps(); err != nil {
		return err
	}

	if err := amo.Push_Contacts(contacts_path, n); err != nil {
		return err
	}

	if err := amo.Push_Leads(leads_path, n); err != nil {
		return err
	}
	return nil
}
