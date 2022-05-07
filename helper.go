package amoimport

import (
	"crypto/sha1"

	"github.com/go-sql-driver/mysql"
)

var mysqlErr *mysql.MySQLError

var sourcesMap = map[string]uint8{}
var usersMap = map[string]uint64{}
var productsMap = map[string]uint32{}
var manufacturersMap = map[string]uint16{}
var stepsMap = map[string]uint8{}
var tagsMap = map[string]uint8{}

func getHash(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	return string(bs)
}
