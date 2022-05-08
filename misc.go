package amoimport

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/hromov/jevelina/auth"
	"github.com/hromov/jevelina/cdb/models"
	"gorm.io/gorm/clause"
)

func (amo *AmoService) Push_Misc(path string, n int) error {
	f, err := os.Open(path)
	if err != nil {
		return errors.New("Unable to read input file " + path + ". Error: " + err.Error())
	}
	defer f.Close()

	r := csv.NewReader(f)

	leadFields = make(map[string]int)

	role, err := auth.GetBaseRole()
	if err != nil {
		return errors.New("Can't get base role from jevelina.auth error: " + err.Error())
	}

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
		respName := leadField(record, "Ответственный")
		amo.saveResponsible(respName, role.ID)

		stepName := leadField(record, "Этап сделки")
		amo.saveMisc(&models.Step{Name: stepName}, stepName)

		prodName := leadField(record, "Товар")
		amo.saveMisc(&models.Product{Name: prodName}, prodName)

		manufName := leadField(record, "Производитель")
		amo.saveMisc(&models.Manufacturer{Name: manufName}, manufName)

		sourceName := leadField(record, "Источник")
		amo.saveMisc(&models.Source{Name: sourceName}, sourceName)
	}

	return nil
}

func (amo *AmoService) saveMisc(m interface{}, name string) {
	if name == "" {
		return
	}

	if _, exist := amo.misc[name]; !exist {
		amo.misc[name] = true
		errorCheck(amo.DB.Create(m).Error, name)
	}
}

func (amo *AmoService) saveResponsible(name string, role uint8) {
	if name == "" {
		return
	}

	if _, exist := amo.misc[name]; !exist {
		amo.misc[name] = true
		email := fmt.Sprintf("email_%d@gmail.com", len(amo.misc))
		//Hash also = email, because hashing just email could look false-safe
		errorCheck(amo.DB.Omit(clause.Associations).Create(
			&models.User{Name: name, Email: email, Hash: email, RoleID: &role},
		).Error, "users")
	}
}

func errorCheck(err error, name string) {
	if err != nil && errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		log.Printf("Can't create item with name %s error: %s", name, err.Error())
	}
}

func (amo *AmoService) LoadMiscsToMaps() error {
	var sources []models.Source
	if err := amo.DB.Find(&sources).Error; err != nil {
		return err
	}
	if len(sources) == 0 {
		log.Println("No sources were found")
	}

	for _, source := range sources {
		amo.sources[source.Name] = source.ID
	}
	var users []models.User
	if err := amo.DB.Find(&users).Error; err != nil {
		return err
	}
	if len(users) == 0 {
		log.Println("No users were found")
	}
	for _, user := range users {
		amo.users[user.Name] = user.ID
	}

	var products []models.Product
	if err := amo.DB.Find(&products).Error; err != nil {
		return err
	}
	if len(users) == 0 {
		log.Println("No products were found")
	}
	for _, item := range products {
		amo.products[item.Name] = item.ID
	}

	var manufs []models.Manufacturer
	if err := amo.DB.Find(&manufs).Error; err != nil {
		return err
	}
	if len(manufs) == 0 {
		log.Println("No manufs were found")
	}
	for _, item := range manufs {
		amo.manufacturers[item.Name] = item.ID
	}

	var steps []models.Step
	if err := amo.DB.Find(&steps).Error; err != nil {
		return err
	}
	if len(manufs) == 0 {
		log.Println("No steps were found")
	}
	for _, item := range steps {
		amo.steps[item.Name] = item.ID
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
