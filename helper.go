package amoimport

import (
	"crypto/sha1"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/hromov/jevelina/cdb/models"
)

var mysqlErr *mysql.MySQLError

func getHash(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	return string(bs)
}

func textToTask(taskText string, parent uint64, responsible *uint64) *models.Task {
	return &models.Task{
		ParentID:      parent,
		Description:   strings.Trim(taskText, ""),
		ResponsibleID: responsible,
		CreatedID:     responsible,
	}
}
