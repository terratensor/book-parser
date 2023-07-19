package main

import (
	"context"
	"github.com/audetv/book-parser/common/app/repos/book"
	"github.com/audetv/book-parser/common/app/repos/paragraph"
	"github.com/audetv/book-parser/common/app/starter"
	"github.com/audetv/book-parser/common/db/sql/pgGormStore"
	"github.com/audetv/book-parser/common/db/sql/pgstore"
	flag "github.com/spf13/pflag"
	"log"
	"os"
	"os/signal"
)

// outputPath путь по которому лежат книги для париснга
var outputPath string

// Default batch size
var batchSize int

// Минимальный размер получаемого после обработки параграфа, указывается в кол-ве символов.
// Значение по умолчанию 800 символов, если указано значение 0, то склейки параграфов не будет
var minParSize int

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	flag.StringVarP(
		&outputPath,
		"output",
		"o",
		"./process/",
		"путь хранения файлов для обработки",
	)
	flag.IntVarP(
		&batchSize,
		"batchSize",
		"b",
		3000,
		"размер пакета по умолчанию (default batch size)",
	)
	flag.IntVarP(&minParSize, "minParSize", "m", 800, "граница минимального размера параграфа в символах, если 0, то без склейки параграфов")

	flag.Parse()

	var bookStore book.BookStore
	var paragraphStore paragraph.ParagraphStore

	storeType := os.Getenv("PARSER_STORE")
	if storeType == "" {
		storeType = "gorm"
	}

	switch storeType {
	case "pg":
		dsn := os.Getenv("PG_DSN")
		if dsn == "" {
			dsn = "postgres://app:secret@localhost:54322/common-library?sslmode=disable"
		}
		pgBookStore, err := pgstore.NewBooks(dsn)
		pgParagraphStore, err := pgstore.NewParagraphs(dsn)
		if err != nil {
			log.Fatal(err)
		}
		defer pgBookStore.Close()
		defer pgParagraphStore.Close()
		bookStore = pgBookStore
		paragraphStore = pgParagraphStore
	case "gorm":
		dsn := os.Getenv("PG_DSN")
		if dsn == "" {
			dsn = "host=localhost user=app password=secret dbname=common-library port=54322 sslmode=disable TimeZone=Europe/Moscow"
		}
		log.Println("подготовка соединения с базой данных")
		pgBookStore, err := pgGormStore.NewBooks(dsn)
		pgParagraphStore, err := pgGormStore.NewParagraphs(dsn)
		log.Println("успешно завершено")

		if err != nil {
			log.Fatal(err)
		}
		bookStore = pgBookStore
		paragraphStore = pgParagraphStore
	default:
		log.Fatal("unknown PARSER_STORE = ", storeType)
	}

	app := starter.NewApp(bookStore, paragraphStore, batchSize, minParSize)

	// читаем все файлы в директории
	files, err := os.ReadDir(outputPath)
	if err != nil {
		log.Fatal(err)
	}

	// итерируемся по списку файлов
	for n, file := range files {
		if file.IsDir() == false {
			// если файл gitignore, то ничего не делаем пропускаем и продолжаем цикл
			if file.Name() == ".gitignore" {
				continue
			}
			app.Parse(ctx, n, file, outputPath)
		}
	}

	log.Println("all files done")
}
