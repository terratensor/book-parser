GOOS=windows GOARCH=amd64 go build -o ./book-parser-pg.exe ./develop/cmd/main.go

### Парсер docx в postgres

Запустить из папки проекта `docker compose up -d`
запуститься база данных postgres и manticore search

- Будет создана локальная БД с именем book-parser
- Имя пользователя app
- Пароль secret
- Порт для подключения к БД 54322

Эти настройки можно увидеть или изменить в файле docker-compose.yml
```
    environment:
      APP_ENV: dev
      POSTGRES_USER: app
      POSTGRES_PASSWORD: secret
      POSTGRES_DB: book-parser
```

Запускаем book-parser-pg.exe файл, папка по умолчанию, где должны быть размещены docx для парсинга — `process`, с помощью параметра `-o` можно указать путь нужный путь к папке, в конце пути обязательно поставить слэш:

`book-parser-pg.exe -o ./mt/`

Будет произведена обработка docx файлов и запись их в таблицы БД:
```
books
    uuid, name, filename, created_at, updated_at, deleted_at

book_paragraphs
    uuid, book_uuid, text, position, created_at, updated_at, deleted_at
    
```

Посмотреть данные в БД можно, например, с помощью программы DBeaver Community

https://dbeaver.io/