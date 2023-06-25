Парсер толстых книг ВП ССССР.

Скопировать файл book-parser-vpsssr.exe в папку с проектом common-lib
Проект должен быть запущен

##### Сборка бинарника, нужно для разработки:

```
GOOS=windows GOARCH=amd64 go build -o ./book-parser-vpsssr.exe ./vpsssr/cmd/main.go
```
