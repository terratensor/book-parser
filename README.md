# book-parser

**Это тестовая начальная версия, ещё многое надо сделать. Парсер не обрабатывает сноски [VII]. Такие сноски попадают в базу как отдельные параграфы. Парсер читает параграфы, и каждый параграф записывает отдельной записью в базу.** 

Для тестирования нужны docker windows, git windows, postman.

- Установить Docker для Windows https://docs.docker.com/desktop/install/windows-install/ 

- Установить git для windows https://git-scm.com/downloads
- При установке можно оставить все по умолчанию

- Установить программу Postman https://www.postman.com/downloads/


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
git clone https://github.com/terratensor/book-parser
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

Запустится новый контейнер с мантикорой, имя будет book-parser (bs-manticore-1)

Скопировать файлы или файл для обработки в папку ./books/VPSSSR/process

Запускаем парсер. Парсер удалит таблицу booksearch в мантикоре и создаст её снова. Если ранее там были обработанные данные (книги), сохранённые после предыдущего запуска парсера, данные будут уничтожены. Далее парсер просканирует папку books/VPSSSR/process/ и обработает по очереди все файлы, которые расположены в этой папке и запишет результат в базу Manticoresearch.

Для запуска парсера книг набрать в консоли Git CMD:

```
book-parser64.exe
```
При необходимости — указать в консоли Git CMD полный путь к этому файлу.

При необходимости можно изменить путь обработки файлов по умолчанию, для этого надо при запуске парсера в командной строке указать опцию -o /путь/к/папке

Например, если указать:
```
book-parser64.exe -o ./books/VPSSSR/DOCX/
```
Будет обработан весь архив толстых книг ВП СССР, находящийся в папке `/books/VPSSSR/DOCX/`

После запуска отобразится листинг параграфов книг, через некоторое время, скорее всего меньше минуты, как парсер обработает книги, можно запустить программу postman

В программе Postman сделать вкладку и в адресе указать:
```
POST localhost:9308/search
```
![image](https://github.com/audetv/book-parser/assets/129882753/aad0a4c2-e213-46ac-a615-4ac98a8a7f82)


Ниже выбрать Body, ниже raw, рядом JSON

В поле ниже скопировать запрос:

```
{
    "index": "booksearch",
    "highlight": {
        "fields": [
            "text"
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

Если надо посмотреть следующие 100 записей, изменить offset, например:
```
"limit": 100
"offset": 100
```

### Опции командной строки

```
-h --help помощь
-o, --output string   путь хранения файлов для обработки (default "./books/VPSSSR/process/")
```
