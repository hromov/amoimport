package amoimport

import (
	"crypto/sha1"
	"encoding/csv"
	"errors"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/hromov/jevelina/cdb/models"
)

var mysqlErr *mysql.MySQLError
var contactFields map[string]int

//key = hash, val = id
var contactsMap map[string]uint64 = map[string]uint64{}

func hashIt(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	return string(bs)
}

func (is *ImportService) Push_Contacts(path string, n int) error {
	f, err := os.Open(path)
	if err != nil {
		return errors.New("Unable to read input file " + path + ". Error: " + err.Error())
	}
	defer f.Close()
	r := csv.NewReader(f)
	contactFields = make(map[string]int)
	for i := 0; i < n; i++ {
		record, err := r.Read()
		// Stop at EOF.
		if err == io.EOF {
			break
		}

		if err != nil {
			panic(err)
		}

		if i == 0 {
			for index, value := range record {
				contactFields[value] = index
			}
			continue
		}
		// Display record.
		// ... Display record length.
		// ... Display all individual elements of the slice.

		// fmt.Println(record)
		// fmt.Println(len(record))
		// for value := range record {
		// 	fmt.Printf(" %d = %v\n", value, record[value])
		// }

		if contact := recordToContact(record); contact != nil {
			responsible := uMap[field(record, "Ответственный")]
			created := uMap[field(record, "Кем создан контакт")]
			source := sMap[field(record, "Источник")]
			if responsible != 0 {
				contact.ResponsibleID = &responsible
			}
			if created != 0 {
				contact.CreatedID = &created
			}
			if source != 0 {
				contact.SourceID = &source
			}
			tags := []models.Tag{}
			for _, tag := range strings.Split(field(record, "Теги"), ",") {
				if _, exist := tagsMap[tag]; exist {
					tags = append(tags, models.Tag{ID: tagsMap[tag]})
				}
			}
			if len(tags) != 0 {
				contact.Tags = tags
			}
			// .Omit(clause.Associations)
			if err := is.DB.Create(contact).Error; err != nil {
				if !errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
					log.Printf("Can't create contact for record # = %d error: %s", i, err.Error())
				} else {
					log.Printf("Can't create contact. Respoonsible = %d, (%+v), created = %d (%+v), source = %d (%+v)", responsible, uMap, created, uMap, source, sMap)
				}
			}

			for _, r := range record[contactFields["Примечание 1"]:contactFields["Примечание 5"]] {
				if r != "" {
					notice := &models.Task{ParentID: contact.ID, Description: strings.Trim(r, ""), ResponsibleID: &responsible, CreatedID: &responsible}
					if err := is.DB.Create(notice).Error; err != nil {
						log.Println(err)
					}
				}
			}
			// } else {
			// 	log.Printf("contacts for record # = %d created: %+v", i, c)
			// }
			//notices 1-5, fullname, contact responsible, records[21:30], records[30:44]
			str := field(record, "Полное имя контакта") + field(record, "Ответственный") + strings.Join(record[contactFields["Рабочий телефон"]:contactFields["Web"]], ",") + strings.Join(record[contactFields["Город"]:contactFields["tid"]], ",")
			// log.Println(str)
			hashed := hashIt(str)
			if _, exist := contactsMap[hashed]; exist {
				log.Println("WTF!!!!!!! contact exist with hash = ", hashed)
				log.Println(contact)
			} else {
				contactsMap[hashed] = contact.ID
			}
		}

	}
	return nil

	// csvReader := csv.NewReader(f)
	// records, err := csvReader.ReadAll()
	// if err != nil {
	// 	return errors.New("Error parsing file: " + err.Error())
	// }

	// return records
}

func field(record []string, name string) string {
	return record[contactFields[name]]
}

func recordToContact(record []string) *models.Contact {
	if len(record) == 0 {
		return nil
	}
	// if len(record) != 43 {
	// 	log.Println("Wrong record schema? len(record) = ", len(record))
	// 	log.Println(record)
	// 	return nil
	// }
	contact := &models.Contact{}
	id, err := strconv.ParseUint(field(record, "ID"), 10, 64)
	if err != nil || id == 0 {
		log.Println("ID parse error: " + err.Error())
		return nil
	}
	contact.ID = id
	if field(record, "Тип") == "контакт" {
		contact.IsPerson = true
	}
	if field(record, "Имя") != "" {
		contact.Name = field(record, "Имя")
	} else {
		contact.Name = field(record, "Полное имя контакта")
	}
	contact.SecondName = field(record, "Фамилия")
	if !contact.IsPerson && field(record, "Название компании") != "" {
		contact.Name = field(record, "Название компании")
	}
	//implement real user by get func
	contact.ResponsibleID = nil
	contact.CreatedID = nil

	const timeForm = "02.01.2006 15:04:05"
	if t, err := time.Parse(timeForm, field(record, "Дата создания контакта")); err == nil {
		contact.CreatedAt = t
	}
	if t, err := time.Parse(timeForm, field(record, "Дата редактирования")); err == nil {
		contact.UpdatedAt = t
	}

	//contact.tags = getTags
	//contact.notices = getNotices record[13:18]

	// Phones start
	dc := regexp.MustCompile(`[^\d|,]`)
	str := dc.ReplaceAllString(strings.Join(record[contactFields["Рабочий email"]:contactFields["Рабочий email"]], ","), "")
	digits := regexp.MustCompile(`(\d){6,13}`)
	// log.Println(str)
	phones := digits.FindAllString(str, -1)
	// log.Println(phones, len(phones))
	switch len(phones) {
	case 0:
		log.Printf("no phones found for contact: %d\n", contact.ID)
		break
	case 1:
		contact.Phone = phones[0]
	default:
		contact.Phone = phones[0]
		contact.SecondPhone = strings.Join(phones[1:], ",")
	}
	// Phones End

	// Email start
	mx := regexp.MustCompile(`[\w-\.]+@([\w-]+\.)+[\w-]{2,4}`)
	emails := mx.FindAllString(strings.Join(record[leadFields["Рабочий email"]:leadFields["Web"]], ","), -1)
	switch len(emails) {
	case 0:
		break
	case 1:
		contact.Email = emails[0]
	default:
		contact.SecondEmail = strings.Join(emails[1:], ",")
	}
	// Email end
	contact.URL = field(record, "Web")
	contact.Address = field(record, "Адрес")
	contact.City = field(record, "Город")

	// implements real source
	contact.SourceID = nil

	contact.Position = field(record, "Должность")

	contact.Analytics.CID = field(record, "cid")
	contact.Analytics.UID = field(record, "uid")
	contact.Analytics.TID = field(record, "tid")

	// log.Printf("all ok: %+v", contact)
	return contact
}

// 0 = ID
//  1 = Тип
//  2 = Полное имя контакта
//  3 = Имя
//  4 = Фамилия
//  5 = Название компании
//  6 = Ответственный
//  7 = Дата создания контакта
//  8 = Кем создан контакт
//  9 = Сделки
//  10 = Дата редактирования
//  11 = Кем редактирован
//  12 = Теги
//  13 = Примечание 1
//  14 = Примечание 2
//  15 = Примечание 3
//  16 = Примечание 4
//  17 = Примечание 5
//  18 = Рабочий телефон
//  19 = Рабочий прямой телефон
//  20 = Мобильный телефон
//  21 = Факс
//  22 = Домашний телефон
//  23 = Другой телефон
//  24 = Рабочий email
//  25 = Личный email
//  26 = Другой email
//  27 = Web
//  28 = Адрес
//  29 = Город
//  30 = Источник
//  31 = Должность
//  32 = Товар
//  33 = Skype
//  34 = ICQ
//  35 = Jabber
//  36 = Google Talk
//  37 = MSN
//  38 = Другой IM
//  39 = Пользовательское соглашение
//  40 = cid
//  41 = uid
//  42 = tid

//  21 = Рабочий телефон
//  22 = Рабочий прямой телефон
//  23 = Мобильный телефон
//  24 = Факс
//  25 = Домашний телефон
//  26 = Другой телефон
//  27 = Рабочий email
//  28 = Личный email
//  29 = Другой email
//  30 = Город
//  31 = Источник
//  32 = Должность
//  33 = Товар
//  34 = Skype
//  35 = ICQ
//  36 = Jabber
//  37 = Google Talk
//  38 = MSN
//  39 = Другой IM
//  40 = Пользовательское соглашение
//  41 = cid
//  42 = uid
//  43 = tid
