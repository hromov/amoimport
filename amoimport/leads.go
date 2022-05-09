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

	brokeLeads, err := os.Create(Broken_leads)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	defer brokeLeads.Close()
	brokenWriter := csv.NewWriter(brokeLeads)

	leadFields = make(map[string]int)
	for i := 0; i < n; i++ {

		record, err := r.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			panic(err)
		}

		if i == 0 {
			recordToLeadFileds(record)
			continue
		}

		if lead := amo.recordToLead(record); lead != nil {
			amo.PushLead(record, lead)
		} else {
			brokenWriter.Write(record)
		}
	}

	brokenWriter.Flush()
	return nil
}

func recordToLeadFileds(record []string) {
	for index, value := range record {
		leadFields[value] = index
	}
}

func (amo *AmoService) PushLead(record []string, lead *models.Lead) {
	if err := amo.DB.Create(lead).Error; err != nil {
		log.Println(err)
	} else {
		taskFields := record[leadFields["Примечание"]:leadFields["Примечание 5"]]
		amo.pushTasks(taskFields, lead.ID, lead.ResponsibleID)
	}

}

func (amo *AmoService) recordToLead(record []string) *models.Lead {
	if len(record) == 0 {
		return nil
	}
	if len(record) != 76 {
		log.Println("Wrong record schema for leads? len(record) = ", len(record))
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
	lead.ContactID = amo.contactIDByLeadRecord(record)
	if lead.ContactID == nil {
		return nil
	}

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

	lead.Analytics.CID = leadField(record, "cid")
	lead.Analytics.UID = leadField(record, "uid")
	lead.Analytics.TID = leadField(record, "tid")
	lead.Analytics.Domain = leadField(record, "domain")

	return lead
}

func (amo *AmoService) contactIDByLeadRecord(record []string) *uint64 {
	stringToHash := leadField(record, "Полное имя контакта") + leadField(record, "Ответственный за контакт") + strings.Join(record[leadFields["Рабочий телефон"]:leadFields["Город"]], ",")
	if contactID, exist := amo.contacts[getHash(stringToHash)]; exist {
		return &contactID
	}
	return nil
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
