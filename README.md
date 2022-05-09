# amoimport

Import from AmoCRM's csv files to [Jevelina](https://github.com/hromov/jevelina)

## How to use

Change these constants according to your needs
```
const (
	leads        = "_import/amocrm_export_leads_2022-04-20.csv"
	contacts     = "_import/amocrm_export_contacts_2022-04-20.csv"
	rowsToImport = 10000
	dsn          = "root:password@tcp(127.0.0.1:3306)/gorm_test?parseTime=True&charset=utf8mb4"
)
```
then just
```
go run .
```
All broken leads and contacts will be in the **broken** folder
