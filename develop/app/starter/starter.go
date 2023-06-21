package starter

import (
	"context"
	"fmt"
	"github.com/audetv/book-parser/develop/app/repos/book"
	"github.com/audetv/book-parser/develop/app/repos/paragraph"
	"github.com/audetv/book-parser/parser/docc"
	"github.com/bwmarrin/snowflake"
	"io"
	"log"
	"os"
	"path/filepath"
	"unicode/utf8"
)

type App struct {
	bs   *book.Books
	ps   *paragraph.Paragraphs
	node *snowflake.Node
}

func NewApp(bookStore book.BookStore, paragraphStore paragraph.ParagraphStore) *App {
	// Создаём новый узел Node с номером 1 для генерации IDs по алгоритму snowflake
	node, err := snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}

	app := &App{
		bs:   book.NewBooks(bookStore),
		ps:   paragraph.NewParagraphs(paragraphStore),
		node: node,
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

	param := 1
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
				Text:     text,
				Position: position,
				Length:   length,
			}

			pars = append(pars, parsedParagraph)

			position++
			param++

			// Записываем пакетам по 2000 параграфов
			if param == 2000 {
				err = app.ps.BulkInsert(ctx, pars, len(pars))
				if err != nil {
					log.Println(err)
				}
				// очищаем slice
				pars = nil
				param = 1
			}
		}
	}

	err = app.ps.BulkInsert(ctx, pars, len(pars))

	log.Printf("%v #%v done", newBook.Filename, n+1)
}
