package starter

import (
	"context"
	"fmt"
	"github.com/audetv/book-parser/common/app/repos/book"
	"github.com/audetv/book-parser/common/app/repos/paragraph"
	"github.com/audetv/book-parser/parser/docc"
	"github.com/google/uuid"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

type App struct {
	bs         *book.Books
	ps         *paragraph.Paragraphs
	batchSize  int
	minParSize int
}

func NewApp(bookStore book.BookStore, paragraphStore paragraph.ParagraphStore, batchSize int, minParSize int) *App {
	app := &App{
		bs:         book.NewBooks(bookStore),
		ps:         paragraph.NewParagraphs(paragraphStore),
		batchSize:  batchSize,
		minParSize: minParSize,
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
	var b strings.Builder

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

			b.WriteString(text)
			// Кол-во символов в параграфе
			length := utf8.RuneCountInString(b.String())

			if app.minParSize != 0 && length < app.minParSize {
				continue
			}

			pars = appendParagraph(b, newBook, position, pars)
			b.Reset()

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

	// Если билдер строки не пустой, записываем оставшийся текст в параграфы и сбрасываем билдер
	if utf8.RuneCountInString(b.String()) > 0 {
		pars = appendParagraph(b, newBook, position, pars)
	}
	b.Reset()

	// Если batchSizeCount меньше batchSize, то записываем оставшиеся параграфы
	if len(pars) > 0 {
		err = app.ps.BulkInsert(ctx, pars, len(pars))
	}

	log.Printf("%v #%v done", newBook.Filename, n+1)
}

func appendParagraph(b strings.Builder, newBook *book.Book, position int, pars paragraph.PrepareParagraphs) paragraph.PrepareParagraphs {
	parsedParagraph := paragraph.Paragraph{
		Uuid:     uuid.New(),
		BookID:   newBook.ID,
		BookName: newBook.Name,
		Text:     b.String(),
		Position: position,
		Length:   utf8.RuneCountInString(b.String()),
	}

	pars = append(pars, parsedParagraph)
	return pars
}
