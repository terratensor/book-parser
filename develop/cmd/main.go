package main

import (
	"context"
	"github.com/audetv/book-parser/develop/app/repos/book"
	"github.com/audetv/book-parser/develop/app/repos/paragraph"
	"github.com/audetv/book-parser/develop/app/starter"
	"github.com/audetv/book-parser/develop/db/sql/pgstore"
	flag "github.com/spf13/pflag"
	"log"
	"os"
	"os/signal"
)

// outputPath путь по которому лежат книги для париснга
var outputPath string

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	flag.StringVarP(
		&outputPath,
		"output",
		"o",
		"./process/",
		"путь хранения файлов для обработки",
	)
	flag.Parse()

	var bookStore book.BookStore
	var paragraphStore paragraph.ParagraphStore

	storeType := os.Getenv("PARSER_STORE")
	if storeType == "" {
		storeType = "pg"
	}

	switch storeType {
	case "pg":
		dsn := os.Getenv("PG_DSN")
		if dsn == "" {
			dsn = "postgres://app:secret@localhost:54322/book-parser?sslmode=disable"
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
	default:
		log.Fatal("unknown PARSER_STORE = ", storeType)
	}

	app := starter.NewApp(bookStore, paragraphStore)

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