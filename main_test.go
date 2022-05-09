package main

import (
	"encoding/csv"
	"os"
	"testing"

	"github.com/hromov/amoimport/amoimport"
	"github.com/hromov/jevelina/cdb"
	"github.com/hromov/jevelina/cdb/models"
)

// const path = "../jevelina/"
const ImportTestDSN = "root:password@tcp(127.0.0.1:3306)/tests_only?charset=utf8mb4&parseTime=True&loc=Local"

type ImportTest struct {
	name          string
	objectType    interface{}
	expectedLen   int
	expectedError error
}

func TestImport(t *testing.T) {
	db, err := cdb.Open(ImportTestDSN)
	if err != nil {
		t.Fatalf("Cant open data base error: %s", err.Error())
	}
	db.DB.Exec("DROP TABLES contacts, contacts_tags, leads, leads_tags, manufacturers, products, roles, sources, steps, tags, tasks, users, files, task_types, transfers, wallets")

	if err = db.Init(); err != nil {
		t.Fatalf("Cant init data base error: %s", err.Error())
	}

	const leads = "test_files/leads_test.csv"
	const contacts = "test_files/contacts_test.csv"
	if err := amoimport.Import(db.DB, leads, contacts, 100); err != nil {
		t.Errorf("Error while importing test data: %s", err.Error())
	}

	tests := []ImportTest{
		{
			name:          "contacts",
			objectType:    models.Contact{},
			expectedLen:   19,
			expectedError: nil,
		},
		{
			name:          "leads",
			objectType:    models.Lead{},
			expectedLen:   19,
			expectedError: nil,
		},
		{
			name:          "users",
			objectType:    models.User{},
			expectedLen:   6,
			expectedError: nil,
		},
		{
			name:          "sources",
			objectType:    models.Source{},
			expectedLen:   2,
			expectedError: nil,
		},
		{
			name:          "products",
			objectType:    models.Product{},
			expectedLen:   3,
			expectedError: nil,
		},
		{
			name:          "manufacturers",
			objectType:    models.Manufacturer{},
			expectedLen:   3,
			expectedError: nil,
		},
		{
			name:          "steps",
			objectType:    models.Step{},
			expectedLen:   5,
			expectedError: nil,
		},
		{
			name:          "tasks",
			objectType:    models.Task{},
			expectedLen:   11,
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filter := models.ListFilter{Limit: 0, Offset: 0}
			switch test.objectType.(type) {
			case models.Contact:
				resp, err := cdb.Contacts().List(filter)
				assertError(t, test.expectedError, err)
				assertLen(t, test.expectedLen, int(resp.Total))
			case models.Lead:
				resp, err := cdb.Leads().List(filter)
				assertError(t, test.expectedError, err)
				assertLen(t, test.expectedLen, int(resp.Total))
			case models.Task:
				resp, err := cdb.Misc().Tasks(filter)
				assertError(t, test.expectedError, err)
				assertLen(t, test.expectedLen, int(resp.Total))
			case models.User:
				users, err := cdb.Misc().Users()
				assertError(t, test.expectedError, err)
				assertLen(t, test.expectedLen, len(users))
			case models.Source:
				sources, err := cdb.Misc().Sources()
				assertError(t, test.expectedError, err)
				assertLen(t, test.expectedLen, len(sources))
			case models.Product:
				products, err := cdb.Misc().Products()
				assertError(t, test.expectedError, err)
				assertLen(t, test.expectedLen, len(products))
			case models.Manufacturer:
				manufs, err := cdb.Misc().Manufacturers()
				assertError(t, test.expectedError, err)
				assertLen(t, test.expectedLen, len(manufs))
			case models.Step:
				steps, err := cdb.Misc().Steps()
				assertError(t, test.expectedError, err)
				assertLen(t, test.expectedLen, len(steps))
			}
		})
	}
	t.Run("Broken_leads", func(t *testing.T) {
		const broken_amount = 1
		f, err := os.Open(amoimport.Broken_leads)
		if err != nil {
			t.Errorf("Can't open %s. Error: %s", amoimport.Broken_leads, err.Error())
		}

		// remember to close the file at the end of the program
		defer f.Close()

		// read csv values using csv.Reader
		csvReader := csv.NewReader(f)
		rows, err := csvReader.ReadAll()
		if err != nil {
			t.Errorf("Can't read %s. Error: %s", amoimport.Broken_leads, err)
		}
		if len(rows) != broken_amount {
			t.Errorf("Expected to have %d broken leads, got: %d", broken_amount, len(rows))
		}
	})
}

func assertError(t *testing.T, expected, real error) {
	if expected != real {
		t.Errorf("Expected to get error: %v, real: %v", expected, real)
	}
}

func assertLen(t *testing.T, expected, real int) {
	if expected != real {
		t.Errorf("Expected len - %v, real - %v", expected, real)
	}
}
