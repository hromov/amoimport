package amoimport

import (
	"encoding/csv"
	"errors"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hromov/jevelina/cdb/models"
)

var contactFields map[string]int

func contactField(record []string, name string) string {
	return record[contactFields[name]]
}

func (amo *AmoService) Push_Contacts(path string, n int) error {
	contactFields = make(map[string]int)
	amo.contacts = make(map[string]uint64)

	f, err := os.Open(path)
	if err != nil {
		return errors.New("Unable to read input file " + path + ". Error: " + err.Error())
	}
	defer f.Close()
	r := csv.NewReader(f)

	brokenFile, err := os.Create(broken_contacts)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	defer brokenFile.Close()
	brokenWriter := csv.NewWriter(brokenFile)

	for i := 0; i < n; i++ {
		record, err := r.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			panic(err)
		}

		if i == 0 {
			recordToContactFileds(record)
			continue
		}

		if contact := amo.recordToContact(record); contact != nil {
			amo.pushContact(record, contact)
		} else {
			brokenWriter.Write(record)
		}
		brokenWriter.Flush()
	}
	return nil
}

func (amo *AmoService) pushContact(record []string, contact *models.Contact) {
	if err := amo.DB.Create(contact).Error; err != nil {
		log.Printf("Can't create contact with ID = %d error: %v", contact.ID, err)
		return
	}
	amo.pushContactsTasks(record, contact)
	amo.contacts[contactHash(record)] = contact.ID
}

func (amo *AmoService) pushContactsTasks(record []string, contact *models.Contact) {
	responsible := amo.users[contactField(record, "Ответственный")]
	for _, taskText := range record[contactFields["Примечание 1"]:contactFields["Примечание 5"]] {
		if taskText != "" {
			amo.DB.Create(textToTask(taskText, contact.ID, &responsible))
		}
	}
}

func recordToContactFileds(record []string) {
	for index, value := range record {
		contactFields[value] = index
	}
}

func contactHash(record []string) string {
	hashString := contactField(record, "Полное имя контакта") + contactField(record, "Ответственный") + strings.Join(record[contactFields["Рабочий телефон"]:contactFields["Web"]], ",")
	return getHash(hashString)
}

func (amo *AmoService) recordToContact(record []string) *models.Contact {
	if len(record) == 0 {
		return nil
	}
	if len(record) != 43 {
		log.Println("Wrong record schema? len(record) = ", len(record))
		return nil
	}
	contact := &models.Contact{}
	id, err := strconv.ParseUint(contactField(record, "ID"), 10, 64)
	if err != nil || id == 0 {
		log.Println("ID parse error: " + err.Error())
		return nil
	}
	contact.ID = id
	if contactField(record, "Тип") == "контакт" {
		contact.IsPerson = true
	}
	if contactField(record, "Имя") != "" {
		contact.Name = contactField(record, "Имя")
	} else {
		contact.Name = contactField(record, "Полное имя контакта")
	}
	contact.SecondName = contactField(record, "Фамилия")
	if !contact.IsPerson && contactField(record, "Название компании") != "" {
		contact.Name = contactField(record, "Название компании")
	}

	responsible := amo.users[contactField(record, "Ответственный")]
	created := amo.users[contactField(record, "Кем создан контакт")]
	source := amo.sources[contactField(record, "Источник")]
	if responsible != 0 {
		contact.ResponsibleID = &responsible
	}
	if created != 0 {
		contact.CreatedID = &created
	}
	if source != 0 {
		contact.SourceID = &source
	}

	const timeForm = "02.01.2006 15:04:05"
	if t, err := time.Parse(timeForm, contactField(record, "Дата создания контакта")); err == nil {
		contact.CreatedAt = t
	}
	if t, err := time.Parse(timeForm, contactField(record, "Дата редактирования")); err == nil {
		contact.UpdatedAt = t
	}

	//contact.tags = getTags
	//contact.notices = getNotices record[13:18]

	// Phones start
	dc := regexp.MustCompile(`[^\d|,]`)
	str := dc.ReplaceAllString(strings.Join(record[contactFields["Рабочий телефон"]:contactFields["Рабочий email"]], ","), "")
	digits := regexp.MustCompile(`(\d){6,13}`)
	phones := digits.FindAllString(str, -1)

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
	emails := mx.FindAllString(strings.Join(record[contactFields["Рабочий email"]:contactFields["Web"]], ","), -1)
	switch len(emails) {
	case 0:
		break
	case 1:
		contact.Email = emails[0]
	default:
		contact.SecondEmail = strings.Join(emails[1:], ",")
	}
	// Email end

	contact.URL = contactField(record, "Web")
	contact.Address = contactField(record, "Адрес")
	contact.City = contactField(record, "Город")

	contact.Position = contactField(record, "Должность")

	contact.Analytics.CID = contactField(record, "cid")
	contact.Analytics.UID = contactField(record, "uid")
	contact.Analytics.TID = contactField(record, "tid")

	tags := []models.Tag{}
	for _, tag := range strings.Split(contactField(record, "Теги"), ",") {
		if _, exist := amo.tags[tag]; exist {
			tags = append(tags, models.Tag{ID: amo.tags[tag]})
		}
	}

	if len(tags) != 0 {
		contact.Tags = tags
	}

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
