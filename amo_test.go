package amoimport

import (
	"os"
	"testing"

	"github.com/hromov/jevelina/base"
)

const path = "../jevelina/"

func TestImport(t *testing.T) {
	dsn, err := os.ReadFile(path + "_keys/db_local")
	if err != nil {
		t.Fatalf("Can't read DB error: %s", err.Error())
	}

	if err := base.Init(string(dsn)); err != nil {
		t.Fatalf("Cant init data base error: %s", err.Error())
	}

	const leads = "import/leads_test.csv"
	const contacts = "import/contacts_test.csv"
	Import(base.GetDB().DB, leads, contacts, 15)

	t.Run("Sources", func(t *testing.T) {
		sources, err := base.GetDB().Misc().Sources()
		if err != nil {
			t.Errorf("Error while Misc().Source(): %s", err)
		}
		if len(sources) == 0 {
			t.Error(("Source len = 0"))
		}
	})
}
