package starter

import (
	"context"
	"fmt"
	"github.com/audetv/book-parser/parser/docc"
	"github.com/audetv/book-parser/vpsssr/app/parser"
	"github.com/audetv/book-parser/vpsssr/app/repos/book"
	"github.com/audetv/book-parser/vpsssr/app/repos/paragraph"
	"io"
	"log"
	"os"
	"path/filepath"
	"unicode/utf8"
)

type App struct {
	bs        *book.Books
	ps        *paragraph.Paragraphs
	batchSize int
}

func NewApp(bookStore book.BookStore, paragraphStore paragraph.ParagraphStore, batchSize int) *App {
	app := &App{
		bs:        book.NewBooks(bookStore),
		ps:        paragraph.NewParagraphs(paragraphStore),
		batchSize: batchSize,
	}
	return app
}

func (app *App) Parse(ctx context.Context, n int, file os.DirEntry, path string) {
	fp := filepath.Clean(fmt.Sprintf("%v%v", path, file.Name()))
	r, err := docc.NewReader(fp)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	// position номер параграфа в индексе
	position := 1

	var filename = file.Name()
	var extension = filepath.Ext(filename)
	var name = filename[0 : len(filename)-len(extension)]

	newBook, err := app.bs.Create(ctx, book.Book{
		Name:     name,
		Filename: filename,
	})
	if err != nil {
		log.Println(err)
	}

	var pars paragraph.PrepareParagraphs

	batchSizeCount := 0
	for {
		text, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		// Если строка не пустая, то записываем в индекс
		if text != "" {

			// Кол-во символов в параграфе
			length := utf8.RuneCountInString(text)

			parsedParagraph := paragraph.Paragraph{
				BookID:   newBook.ID,
				BookName: newBook.Name,
				Text:     text,
				Position: position,
				Length:   length,
			}

			pars = append(pars, parsedParagraph)

			position++
			batchSizeCount++

			// Записываем пакетам по batchSize параграфов
			if batchSizeCount == app.batchSize-1 {
				err = app.ps.BulkInsert(ctx, pars, len(pars))
				if err != nil {
					log.Printf("log bulk insert error query: %v \r\n", err)
				}
				// очищаем slice
				pars = nil
				batchSizeCount = 0
			}
		}
	}

	// Если batchSizeCount меньше batchSize, то записываем оставшиеся параграфы
	if len(pars) > 0 {
		err = app.ps.BulkInsert(ctx, pars, len(pars))
	}

	log.Printf("%v #%v done", newBook.Filename, n+1)
}

func (app *App) Process(ctx context.Context, books parser.Books) {

	for n, pb := range books {
		newBook, err := app.bs.Create(ctx, book.Book{
			Name:     pb.Name,
			Filename: pb.Filename,
		})
		if err != nil {
			log.Println(err)
		}

		var pars paragraph.PrepareParagraphs
		batchSizeCount := 0
		for _, pp := range pb.Paragraphs {

			// Кол-во символов в параграфе
			length := utf8.RuneCountInString(pp.Text)

			parsedParagraph := paragraph.Paragraph{
				BookID:   newBook.ID,
				BookName: newBook.Name,
				Text:     pp.Text,
				Position: pp.Position,
				Length:   length,
			}

			pars = append(pars, parsedParagraph)

			batchSizeCount++

			// Записываем пакетам по batchSize параграфов
			if batchSizeCount == app.batchSize-1 {
				err = app.ps.BulkInsert(ctx, pars, len(pars))
				if err != nil {
					log.Printf("log bulk insert error query: %v \r\n", err)
				}
				// очищаем slice
				pars = nil
				batchSizeCount = 0
			}
		}

		// Если batchSizeCount меньше batchSize, то записываем оставшиеся параграфы
		if len(pars) > 0 {
			err = app.ps.BulkInsert(ctx, pars, len(pars))
		}

		log.Printf("%v #%v done", newBook.Filename, n+1)
	}
}
