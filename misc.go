package amoimport

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/hromov/jevelina/cdb/models"
	"gorm.io/gorm/clause"
)

var sMap = map[string]uint8{}
var uMap = map[string]uint64{}
var pMap = map[string]uint32{}
var mMap = map[string]uint16{}
var stepsMap = map[string]uint8{}
var tagsMap = map[string]uint8{}

func (is *ImportService) Push_Misc(path string, n int) error {
	f, err := os.Open(path)
	if err != nil {
		return errors.New("Unable to read input file " + path + ". Error: " + err.Error())
	}
	defer f.Close()

	r := csv.NewReader(f)
	misc := map[string]int{}

	role := &models.Role{Role: "Admin"}
	if err := is.DB.Create(role).Error; err != nil {
		if !errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			log.Printf("Can't create admin role error: %s", err.Error())
		}
	}

	role = &models.Role{Role: "User"}
	if err := is.DB.Create(role).Error; err != nil {
		if !errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			log.Printf("Can't create user role error: %s", err.Error())
		}
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
		if _, exist := misc[record[3]]; !exist && record[3] != "" {
			misc[record[3]] = -1
			email := fmt.Sprintf("email_%d@gmail.com", i)
			//Hash also = email, because hashing just email could be dangerous
			if err := is.DB.Omit(clause.Associations).Create(&models.User{Name: record[3], Email: email, Hash: email, RoleID: &role.ID}).Error; err != nil {
				if !errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
					log.Printf("Can't create user for record # = %d error: %s", i, err.Error())
				}
			}
		}
		if _, exist := misc[record[15]]; !exist && record[15] != "" {
			misc[record[15]] = -1
			if err := is.DB.Create(&models.Step{Name: record[15]}).Error; err != nil {
				if !errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
					log.Printf("Can't create step for record # = %d error: %s", i, err.Error())
				}
			}
		}
		if _, exist := misc[record[69]]; !exist && record[69] != "" {
			misc[record[69]] = -1
			if err := is.DB.Create(&models.Product{Name: record[69]}).Error; err != nil {
				if !errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
					log.Printf("Can't create product for record # = %d error: %s", i, err.Error())
				}
			}
		}
		if _, exist := misc[record[70]]; !exist && record[70] != "" {
			misc[record[70]] = -1
			if err := is.DB.Create(&models.Manufacturer{Name: record[70]}).Error; err != nil {
				if !errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
					log.Printf("Can't create manufacturer for record # = %d error: %s", i, err.Error())
				}
			}
		}
		if _, exist := misc[record[31]]; !exist && record[31] != "" {
			misc[record[31]] = -1
			if err := is.DB.Create(&models.Source{Name: record[31]}).Error; err != nil {
				if !errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
					log.Printf("Can't create source for record # = %d error: %s", i, err.Error())
				}
			}
		}
		for _, tag := range strings.Split(record[9], ",") {
			if _, exist := misc[tag]; !exist && tag != "" {
				misc[tag] = -1
				if err := is.DB.Create(&models.Tag{Name: tag}).Error; err != nil {
					if !errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
						log.Printf("Can't create source for record # = %d error: %s", i, err.Error())
					}
				}
			}
		}

	}
	var sources []*models.Source
	if err := is.DB.Find(&sources).Error; err != nil {
		return err
	}
	if len(sources) == 0 {
		log.Println("No sources were found")
	}

	for _, source := range sources {
		sMap[source.Name] = source.ID
	}
	var users []*models.User
	if err := is.DB.Find(&users).Error; err != nil {
		return err
	}
	if len(users) == 0 {
		log.Println("No users were found")
	}
	for _, user := range users {
		uMap[user.Name] = user.ID
	}

	var products []*models.Product
	if err := is.DB.Find(&products).Error; err != nil {
		return err
	}
	if len(users) == 0 {
		log.Println("No products were found")
	}
	for _, item := range products {
		pMap[item.Name] = item.ID
	}

	var manufs []*models.Manufacturer
	if err := is.DB.Find(&manufs).Error; err != nil {
		return err
	}
	if len(manufs) == 0 {
		log.Println("No manufs were found")
	}
	for _, item := range manufs {
		mMap[item.Name] = item.ID
	}

	var steps []*models.Step
	if err := is.DB.Find(&steps).Error; err != nil {
		return err
	}
	if len(manufs) == 0 {
		log.Println("No steps were found")
	}
	for _, item := range steps {
		stepsMap[item.Name] = item.ID
	}

	var tags []*models.Tag
	if err := is.DB.Find(&tags).Error; err != nil {
		return err
	}
	if len(tags) == 0 {
		log.Println("No steps were found")
	}
	for _, item := range tags {
		tagsMap[item.Name] = item.ID
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
