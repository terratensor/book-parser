### Парсер docx в postgres

Запустить из папки проекта `docker compose up --build -d`
запустится база данных postgres и manticore search

- Будет создана локальная БД с именем common-library
- Имя пользователя app
- Пароль secret
- Порт для подключения к БД 54322

Эти настройки можно увидеть или изменить в файле docker-compose.yml
```
    environment:
      APP_ENV: dev
      POSTGRES_USER: app
      POSTGRES_PASSWORD: secret
      POSTGRES_DB: common-library
```

Запускаем book-parser-pg.exe файл, папка по умолчанию, где должны быть размещены docx для парсинга — `process`, с помощью параметра `-o` можно указать путь нужный путь к папке, в конце пути обязательно поставить слэш:

`book-parser-pg.exe -o ./militera/mt/`

Будет произведена обработка docx файлов и запись их в таблицы БД:
```
books
    id, name, filename, created_at, updated_at, deleted_at

book_paragraphs
    id, book_id, text, position, length, created_at, updated_at, deleted_at 
```

Посмотреть данные в БД можно, например, с помощью программы DBeaver Community

https://dbeaver.io/

Остановка БД и мантикоры

`docker-compose down --remove-orphans`

Запустить из папки проекта 

`docker compose up --build -d`

### Первый запуск manticore indexer, если еще не создана таблица common_library

```
docker exec -it book-parser-manticore indexer common_library
```

### Повторный запуск manticore indexer, если таблица существует, для переиндексации

```
docker exec -it book-parser-manticore indexer common_library --rotate
```

#### Бэкап postgres БД:
```
docker exec -it book-parser-postgres bash

pg_dump --dbname=book-parser --username=app --host=postgres-book-parser | gzip -9 > book-parser-backup-filename.gz
```

Скопировать файл бэкапа в контейнер докера для восстановления БД, распаковать из gz, запустить загрузку бэкапа в БД

```
cp book-parser-backup-filename.gz book-parser-postgres:app/book-parser-backup-filename.gz

gzip -d book-parser-backup-filename.gz

psql -U app -d lib < book-parser-backup-filename.sql
```

#### Остановка и удаление БД. ВНИМАНИЕ ДАННЫЕ БУДУТ УДАЛЕНЫ.

`docker compose down -v --remove-orphans`

`./book-parser-common-dev-0.3.0.exe -m 1000 -p 1800 -x 3000 -o ./m/`

### Опции командной строки

```
-h --help помощь
-b, --batchSize int    размер пакета по умолчанию (default batch size) (default 3000)
-d, --dev              подробный вывод служебной информации об обработке параграфов в лог консоли
-x, --maxParSize int   граница максимального размера параграфа в символах (default 3500)
-m, --minParSize int   граница минимального размера параграфа в символах (default 300)
-p, --optParSize int   граница оптимального размера параграфа в символах (default 1800)
-o, --output string    путь хранения файлов для обработки (default "./process/")
```

##### Сборка бинарника, нужно для разработки:

`GOOS=windows GOARCH=amd64 go build -o ./book-parser-gorm.exe ./common/cmd/main.go`
