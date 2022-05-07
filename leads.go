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

func Get_Contact_ID(record []string) *uint64 {
	//notices 1-5, fullname, contact responsible, records[21:30], records[30:44]
	str := record[17] + record[19] + strings.Join(record[21:30], ",") + strings.Join(record[30:44], ",")
	// log.Println(str)
	hashed := hashIt(str)
	if _, exist := contactsMap[hashed]; !exist {
		log.Println("WTF!!!!!!! can'f find contact for lead = ", str)
		return nil
	}
	r := contactsMap[hashed]
	return &r
}

func (is *ImportService) Push_Leads(path string, n int) error {
	f, err := os.Open(path)
	if err != nil {
		return errors.New("Unable to read input file " + path + ". Error: " + err.Error())
	}
	defer f.Close()

	r := csv.NewReader(f)
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

		if lead := recordToLead(record); lead != nil {

			responsible := uMap[record[3]]
			created := uMap[record[5]]
			source := sMap[record[31]]

			prod := pMap[record[69]]
			manuf := mMap[record[70]]
			step := stepsMap[record[15]]
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
			for _, tag := range strings.Split(record[9], ",") {
				if _, exist := tagsMap[tag]; exist {
					tags = append(tags, models.Tag{ID: tagsMap[tag]})
				}
			}
			if len(tags) != 0 {
				lead.Tags = tags
			}

			if err := is.DB.Create(lead).Error; err != nil {
				if !errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
					log.Printf("Can't create lead for record # = %d error: %s", i, err.Error())
				} else {
					log.Println(err)
				}
			}

			for _, r := range record[10:15] {
				if r != "" {
					notice := &models.Task{ParentID: lead.ID, Description: strings.Trim(r, ""), ResponsibleID: &responsible, CreatedID: &responsible}
					if err = is.DB.Create(notice).Error; err != nil {
						log.Println(err)
					}
				}
			}
			// } else {
			// 	log.Printf("lead for record # = %d created: %+v", i, c)
			// }
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

func recordToLead(record []string) *models.Lead {
	if len(record) == 0 {
		return nil
	}
	if len(record) != 76 {
		log.Println("Wrong record schema for leads? len(record) = ", len(record))
		log.Println(record)
		return nil
	}
	lead := &models.Lead{}
	id, err := strconv.ParseUint(record[0], 10, 64)
	if err != nil || id == 0 {
		log.Println("ID parse error: " + err.Error())
		return nil
	}
	lead.ID = id
	lead.Name = record[1]
	budget, err := strconv.ParseUint(record[2], 10, 32)
	if err == nil {
		lead.Budget = uint32(budget)
	}
	lead.ContactID = Get_Contact_ID(record)
	// lead.ContactID = nil

	//responsible and created from contacts goes here
	//implement real user by get func
	lead.ResponsibleID = nil
	lead.CreatedID = nil

	const timeForm = "02.01.2006 15:04:05"
	if t, err := time.Parse(timeForm, record[4]); err == nil {
		lead.CreatedAt = t
	}

	if t, err := time.Parse(timeForm, record[6]); err == nil {
		lead.UpdatedAt = t
	}
	if t, err := time.Parse(timeForm, record[8]); err == nil {
		lead.ClosedAt = &t
	}
	// lead.ClosedAt, _ = time.Parse(timeForm, record[8])
	// createdTime := strings.ReplaceAll(record[4], ".", "-")
	// created, err := time.Parse(timeForm, createdTime)
	// if err == nil {
	// 	lead.CreatedAt = created
	// }

	// updatedTime := strings.ReplaceAll(record[6], ".", "-")
	// updated, err := time.Parse(timeForm, updatedTime)
	// if err == nil {
	// 	lead.UpdatedAt = updated
	// }

	// closedTime := strings.ReplaceAll(record[8], ".", "-")
	// closed, err := time.Parse(timeForm, closedTime)
	// if err == nil {
	// 	lead.ClosedAt = &closed
	// }

	//tags record[9]
	//genereate from record[15]
	lead.StepID = nil
	lead.ProductID = nil
	lead.ManufacturerID = nil

	lead.Analytics.CID = record[71]
	lead.Analytics.UID = record[72]
	lead.Analytics.TID = record[73]
	// // implements real source record[74]
	// contact.SourceID = nil
	lead.Analytics.Domain = record[75]

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
