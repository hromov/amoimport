package amoimport

import (
	"encoding/csv"
	"errors"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hromov/jevelina/cdb/models"
)

var leadFields map[string]int

const broken_leads = "broken_leads.csv"

func leadField(record []string, name string) string {
	return record[leadFields[name]]
}

func (amo *AmoService) Push_Leads(path string, n int) error {
	f, err := os.Open(path)
	if err != nil {
		return errors.New("Unable to read input file " + path + ". Error: " + err.Error())
	}
	defer f.Close()

	r := csv.NewReader(f)

	csvFile, err := os.Create(broken_leads)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	defer csvFile.Close()
	csvwriter := csv.NewWriter(csvFile)

	leadFields = make(map[string]int)
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
			//in case we did something with misk let's init it one more time
			for index, value := range record {
				leadFields[value] = index
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

		if lead := amo.recordToLead(record); lead != nil {

			responsible := amo.users[leadField(record, "Ответственный")]
			created := amo.users[leadField(record, "Кем создана сделка")]
			source := amo.sources[leadField(record, "Источник")]

			prod := amo.products[leadField(record, "Товар")]
			manuf := amo.manufacturers[leadField(record, "Производитель")]
			step := amo.steps[leadField(record, "Этап сделки")]
			if responsible != 0 {
				lead.ResponsibleID = &responsible
			}
			if created != 0 {
				lead.CreatedID = &created
			}
			if source != 0 {
				lead.SourceID = &source
			}
			if prod != 0 {
				lead.ProductID = &prod
			}
			if manuf != 0 {
				lead.ManufacturerID = &manuf
			}
			if step != 0 {
				lead.StepID = &step
			}
			tags := []models.Tag{}
			for _, tag := range strings.Split(leadField(record, "Теги"), ",") {
				if _, exist := amo.tags[tag]; exist {
					tags = append(tags, models.Tag{ID: amo.tags[tag]})
				}
			}
			if len(tags) != 0 {
				lead.Tags = tags
			}

			if err := amo.DB.Create(lead).Error; err != nil {
				if !errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
					log.Printf("Can't create lead for record # = %d error: %s", i, err.Error())
				} else {
					log.Println(err)
				}
			}

			for _, r := range record[leadFields["Примечание"]:leadFields["Примечание 5"]] {
				if r != "" {
					notice := &models.Task{ParentID: lead.ID, Description: strings.Trim(r, ""), ResponsibleID: &responsible, CreatedID: &responsible}
					if err = amo.DB.Create(notice).Error; err != nil {
						log.Println(err)
					}
				}
			}
			// } else {
			// 	log.Printf("lead for record # = %d created: %+v", i, c)
			// }
		} else {
			_ = csvwriter.Write(record)
		}
	}

	csvwriter.Flush()
	return nil

	// csvReader := csv.NewReader(f)
	// records, err := csvReader.ReadAll()
	// if err != nil {
	// 	return errors.New("Error parsing file: " + err.Error())
	// }

	// return records
}

func (amo *AmoService) recordToLead(record []string) *models.Lead {
	if len(record) == 0 {
		return nil
	}
	if len(record) != 76 {
		log.Println("Wrong record schema for leads? len(record) = ", len(record))
		// log.Println(record)
		return nil
	}
	lead := &models.Lead{}
	id, err := strconv.ParseUint(leadField(record, "ID"), 10, 64)
	if err != nil || id == 0 {
		log.Println("ID parse error: " + err.Error())
		return nil
	}
	lead.ID = id
	lead.Name = leadField(record, "Название сделки")
	budget, err := strconv.ParseUint(leadField(record, "Бюджет"), 10, 32)
	if err == nil {
		lead.Budget = uint32(budget)
	}
	// contactToHash := contactField(record, "Полное имя контакта") + contactField(record, "Ответственный") + strings.Join(record[contactFields["Рабочий телефон"]:contactFields["Web"]], ",")
	// log.Panicln(contactToHash)
	stringToHash := leadField(record, "Полное имя контакта") + leadField(record, "Ответственный за контакт") + strings.Join(record[leadFields["Рабочий телефон"]:leadFields["Город"]], ",")
	if contactID, exist := amo.contacts[getHash(stringToHash)]; exist {
		lead.ContactID = &contactID
	} else {
		// log.Printf("no contact found for lead: %+v", lead)
		return nil
	}

	//responsible and created from contacts goes here
	//implement real user by get func
	lead.ResponsibleID = nil
	lead.CreatedID = nil

	const timeForm = "02.01.2006 15:04:05"
	if t, err := time.Parse(timeForm, leadField(record, "Дата создания сделки")); err == nil {
		lead.CreatedAt = t
	}

	if t, err := time.Parse(timeForm, leadField(record, "Дата редактирования")); err == nil {
		lead.UpdatedAt = t
	}
	if t, err := time.Parse(timeForm, leadField(record, "Дата закрытия")); err == nil {
		lead.ClosedAt = &t
	}

	//tags record[9]
	//genereate from record[15]
	lead.StepID = nil
	lead.ProductID = nil
	lead.ManufacturerID = nil

	lead.Analytics.CID = leadField(record, "cid")
	lead.Analytics.UID = leadField(record, "uid")
	lead.Analytics.TID = leadField(record, "tid")
	// // implements real source record[74]
	// contact.SourceID = nil
	lead.Analytics.Domain = leadField(record, "domain")

	// log.Printf("all ok: %+v", lead)
	return lead
}

//  0 = ID
//  1 = Название сделки
//  2 = Бюджет
//  3 = Ответственный
//  4 = Дата создания сделки
//  5 = Кем создана сделка
//  6 = Дата редактирования
//  7 = Кем редактирована
//  8 = Дата закрытия
//  9 = Теги
//  10 = Примечание
//  11 = Примечание 2
//  12 = Примечание 3
//  13 = Примечание 4
//  14 = Примечание 5
//  15 = Этап сделки
//  16 = Воронка
//  17 = Полное имя контакта
//  18 = Компания контакта
//  19 = Ответственный за контакт
//  20 = Компания
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
//  44 = utm_source
//  45 = utm_medium
//  46 = utm_campaign
//  47 = utm_term
//  48 = utm_content
//  49 = utm_referrer
//  50 = _ym_uid
//  51 = _ym_counter
//  52 = roistat
//  53 = referrer
//  54 = openstat_service
//  55 = openstat_campaign
//  56 = openstat_ad
//  57 = openstat_source
//  58 = from
//  59 = gclientid
//  60 = gclid
//  61 = yclid
//  62 = fbclid
//  63 = GOOGLE_ID
//  64 = roistat
//  65 = KEYWORD
//  66 = ADV_CAMP
//  67 = TRAF_TYPE
//  68 = TRAF_SRC
//  69 = Товар
//  70 = Производитель
//  71 = cid
//  72 = uid
//  73 = tid
//  74 = Источник
//  75 = domain
