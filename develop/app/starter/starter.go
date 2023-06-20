package starter

import (
	"context"
	"fmt"
	"github.com/audetv/book-parser/develop/app/repos/book"
	"github.com/audetv/book-parser/develop/app/repos/paragraph"
	"github.com/audetv/book-parser/parser/docc"
	"github.com/google/uuid"
	"io"
	"log"
	"os"
	"path/filepath"
)

type App struct {
	bs *book.Books
	ps *paragraph.Paragraphs
}

func NewApp(bookStore book.BookStore, paragraphStore paragraph.ParagraphStore) *App {
	app := &App{
		bs: book.NewBooks(bookStore),
		ps: paragraph.NewParagraphs(paragraphStore),
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

	// Генерируем UUID
	ID := uuid.New()
	var filename = file.Name()
	var extension = filepath.Ext(filename)
	var name = filename[0 : len(filename)-len(extension)]

	newBook, err := app.bs.Create(ctx, book.Book{
		ID:       ID,
		Name:     filename,
		Filename: name,
	})
	if err != nil {
		log.Println(err)
	}

	for {
		text, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		// Если строка не пустая, то записываем в индекс
		if text != "" {
			parsedParagraph := paragraph.Paragraph{
				ID:       uuid.New(),
				BookID:   newBook.ID,
				Text:     text,
				Position: position,
			}

			app.ps.Create(ctx, &parsedParagraph)
			position++
		}
	}

	log.Printf("%v #%v done", newBook.Filename, n+1)
}
