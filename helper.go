package amoimport

import (
	"crypto/sha1"
	"strings"

	"github.com/go-sql-driver/mysql"
)

var mysqlErr *mysql.MySQLError

func getHash(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	return string(bs)
}

func (amo *AmoService) Get_Contact_ID(record []string) *uint64 {
	//notices 1-5, fullname, contact responsible, records[21:30], records[30:44]
	str := leadField(record, "Полное имя контакта") + leadField(record, "Ответственный за контакт") + strings.Join(record[leadFields["Рабочий телефон"]:leadFields["utm_source"]], ",")
	// log.Println(str)
	hashed := getHash(str)
	if _, exist := amo.contacts[hashed]; !exist {
		//TODO: put them in separate file and not into the base
		// log.Println("WTF!!!!!!! can'f find contact for lead = ", str)
		return nil
	}
	r := amo.contacts[hashed]
	return &r
}
