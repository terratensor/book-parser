package main

import (
	"context"
	"fmt"
	"github.com/audetv/book-parser/common/app/repos/book"
	"github.com/audetv/book-parser/common/app/repos/paragraph"
	"github.com/audetv/book-parser/common/app/starter"
	"github.com/audetv/book-parser/common/app/workerpool"
	"github.com/audetv/book-parser/common/db/sql/pgGormStore"
	"github.com/audetv/book-parser/common/db/sql/pgstore"
	flag "github.com/spf13/pflag"
	"log"
	"os"
	"os/signal"
	"time"
)

// outputPath путь по которому лежат книги для париснга
var outputPath string

// Default batch size
var batchSize int

// Минимальный размер получаемого после обработки параграфа, указывается в кол-ве символов.
// Значение по умолчанию 800 символов, если указано значение 0, то склейки параграфов не будет
var minParSize int
var optParSize int
var maxParSize int

var devMode bool

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
	flag.IntVarP(&minParSize, "minParSize", "m", 300, "граница минимального размера параграфа в символах, если 0, то без склейки параграфов")
	flag.IntVarP(&optParSize, "optParSize", "p", 1800, "граница оптимального размера параграфа в символах, если 0, то без склейки параграфов")
	flag.IntVarP(&maxParSize, "maxParSize", "x", 3500, "граница максимального размера параграфа в символах, если 0, то без склейки параграфов")
	flag.BoolVarP(&devMode, "dev", "d", false, "подробный вывод служебной информации об обработке параграфов в лог консоли")

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

	app := starter.NewApp(bookStore, paragraphStore, batchSize, minParSize, optParSize, maxParSize, devMode)

	// читаем все файлы в директории
	files, err := os.ReadDir(outputPath)
	if err != nil {
		log.Fatal(err)
	}

	// Срез ошибок полученных при обработке файлов
	var errors []string

	var allTask []*workerpool.Task

	for n, file := range files {
		if file.IsDir() == false {

			// если файл gitignore, то ничего не делаем пропускаем и продолжаем цикл
			if file.Name() == ".gitignore" {
				continue
			}

			task := workerpool.NewTask(func(data interface{}) error {

				fmt.Printf("Task %v processed\n", file.Name())

				err = app.Parse(ctx, n, file, outputPath)
				if err != nil {
					return err
				}
				return nil
			}, file)
			allTask = append(allTask, task)

			errors = append(errors, fmt.Sprintln(err))
		}
	}
	defer duration(track("Обработка завершена за "))
	pool := workerpool.NewPool(allTask, 12)
	pool.Run()

	saveErrors(errors)
	log.Println("all files done")
}

func saveErrors(errors []string) {
	if len(errors) > 0 {
		// Создание файла для записи ошибок при обработке
		currentTime := time.Now()
		logfile := fmt.Sprintf("./%v_error_log.txt", currentTime.Format("15-04-05_02012006"))

		f, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()

		for _, errorFile := range errors {
			data := []byte(fmt.Sprint(errorFile))
			f.Write(data)
		}
	}
}

func track(msg string) (string, time.Time) {
	return msg, time.Now()
}

func duration(msg string, start time.Time) {
	log.Printf("%v: %v\n", msg, time.Since(start))
}
