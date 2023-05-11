# book-parser

Для тестирования нужны docker windows, git windows, postman.

- Установить git для windows https://git-scm.com/downloads
- При установке можно оставить все по умолчанию


После установки, нажать меню пуск, найти git консоль Git CMD, запустить

- Создать папку, например c:\Users\<username>\booksearch
  , набрать в консоли Git CMD
```
mkdir ./booksearch
```

Выбрать созданную папку, набрать в консоли Git CMD: 
```
cd ./booksearch
```

Клонировать репозиторий audetv/book-parser, набрать в консоли Git CMD:
```
git clone https://github.com/audetv/book-parser.git
```
Увидите текст, примерно такой:
```
Cloning into 'book-parser'...
remote: Enumerating objects: 134, done.
remote: Counting objects: 100% (35/35), done.
remote: Compressing objects: 100% (28/28), done.
remote: Total 134 (delta 5), reused 17 (delta 4), pack-reused 99
Receiving objects:  98% (132/134), 31.49 MiB | 1.84 MiB/s
Receiving objects: 100% (134/134), 32.22 MiB | 2.66 MiB/s, done.
Resolving deltas: 100% (7/7), done.
```

После набрать в консоли Git CMD
```
cd ./book-parser
```

Далее запустить windows docker (через меню пуск)

После запуска проверить, что нет запущенных контейнеров с мантикорой (меню containers), если есть остановить их или удалить, иначе будет конфликт доступа к портам.

В консоли Git CMD, где ранее создали папку и клонировали репозиторий для запуска контейнера с мантикорой набрать:

```
docker compose up -d
```

Запуститься новый контейнер с мантикорой, имя будет book-parser (bs-manticore-1)

Для запуска парсера книг набрать в консоли Git CMD:

```
./book-parser64.exe
```

В консоли будет листинг (много строк) параграфов книг, через некоторое время, скорее всего меньше минуты, как парсер обработает книги, можно запустить программу postman

В постмане сделать вкладку и в адресе указать:
```
POST localhost:9308/search
```

Ниже выбрать Body, ниже raw, рядом JSON

В поле ниже скопировать запрос:

```
{
    "index": "booksearch",
    "highlight": {
        "fields": [
            "content"
        ],
        "limit": 0,
        "no_match_size": 0
    },
    "limit": 100,
    "offset": 0,
    "query": {
        "bool": {
            "must": [
                {
                    "match_phrase": { "_all" : "пфу"}
                }
            ]
        }
    }
}
```
В match_phrase, вместо "пфу", можно набрать любой поисковый запрос в кавычках, например "мера"

"limit" - кол-во записей на странице
"offset" - смещение от начала.

Если надо посмотреть следйующие 100 записей, изменить offset, например:
```
"limit": 100
"offset": 100
```
