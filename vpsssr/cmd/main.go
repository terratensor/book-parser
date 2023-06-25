package main

import (
	"context"
	"github.com/audetv/book-parser/vpsssr/app/parser"
	"github.com/audetv/book-parser/vpsssr/app/repos/book"
	"github.com/audetv/book-parser/vpsssr/app/repos/paragraph"
	"github.com/audetv/book-parser/vpsssr/app/starter"
	"github.com/audetv/book-parser/vpsssr/db/sql/pgGormStore"
	flag "github.com/spf13/pflag"
	"log"
	"os"
	"os/signal"
)

// outputPath путь по которому лежат книги для париснга
var outputPath string

// Default batch size
var batchSize int

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

	flag.Parse()

	var bookStore book.BookStore
	var paragraphStore paragraph.ParagraphStore

	storeType := os.Getenv("PARSER_STORE")
	if storeType == "" {
		storeType = "gorm"
	}

	switch storeType {
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

	app := starter.NewApp(bookStore, paragraphStore, batchSize)

	books := parser.Parse(outputPath)
	app.Process(ctx, books)

	log.Println("all files done")
}
