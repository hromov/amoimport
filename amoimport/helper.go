package amoimport

import (
	"crypto/sha1"
	"errors"
	"log"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/hromov/jevelina/cdb/models"
)

var mysqlErr *mysql.MySQLError

func (amo *AmoService) pushTasks(record []string, parent uint64, responsible *uint64) {
	for _, taskText := range record {
		if taskText != "" {
			amo.DB.Create(textToTask(taskText, parent, responsible))
		}
	}
}

func textToTask(taskText string, parent uint64, responsible *uint64) *models.Task {
	return &models.Task{
		ParentID:      parent,
		Description:   strings.Trim(taskText, ""),
		ResponsibleID: responsible,
		CreatedID:     responsible,
	}
}

func errorCheck(err error, name string) {
	if err != nil && errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		log.Printf("Can't create item with name %s error: %s", name, err.Error())
	}
}

func getHash(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	return string(bs)
}
