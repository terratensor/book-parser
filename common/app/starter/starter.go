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
	optParSize int
	maxParSize int
}

func NewApp(
	bookStore book.BookStore,
	paragraphStore paragraph.ParagraphStore,
	batchSize int,
	minParSize int,
	optParSize int,
	maxParSize int,
) *App {
	app := &App{
		bs:         book.NewBooks(bookStore),
		ps:         paragraph.NewParagraphs(paragraphStore),
		batchSize:  batchSize,
		minParSize: minParSize,
		optParSize: optParSize,
		maxParSize: maxParSize,
	}
	return app
}

func (app *App) Parse(ctx context.Context, n int, file os.DirEntry, path string) error {
	fp := filepath.Clean(fmt.Sprintf("%v%v", path, file.Name()))

	var filename = file.Name()
	var extension = filepath.Ext(filename)
	var name = filename[0 : len(filename)-len(extension)]

	r, err := docc.NewReader(fp)
	if err != nil {
		str := fmt.Sprintf("%v, %v", filename, err)
		log.Println(str)
		return fmt.Errorf(str)
	}
	defer r.Close()

	// position номер параграфа в индексе
	position := 1

	newBook, err := app.bs.Create(ctx, book.Book{
		Name:     name,
		Filename: filename,
	})
	if err != nil {
		log.Println(err)
	}

	var pars paragraph.PrepareParagraphs
	var b strings.Builder
	var bufBuilder strings.Builder

	batchSizeCount := 0
	for {
		// Используем select для выхода по истечении контекста, прерывание выполнения ctrl+c
		select {
		case <-ctx.Done():
			log.Printf("ctx done: %v", ctx.Err())
			return ctx.Err()
		default:
		}

		text, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			str := fmt.Sprintf("%v, %v", filename, err)
			log.Println(str)
			return fmt.Errorf(str)
		}

		// Скорее всего надо сделать здесь ещё цикл и в нем делать проверки
		// Для того чтобы разбитый параграф возвращал оптимальную часть и остаток,
		// которые надо проверить, оптимальную часть записать к pars, а остаток проверить на длинный параграф,
		// и если параграф длинный то в этом цикле крутить пока не обработает весь длинный параграф,
		// только после этого уже начать обрабатывать текущий text, только после того как предыдущий длинный параграф весь будет разбит
		// в итерациях этого цикла, обработан и записан в pars

		// Если в билдер-буфере есть записанный параграф, то записываем его в обычны билдер b и очищаем билдер-буфер
		if utf8.RuneCountInString(bufBuilder.String()) > 0 {
			if utf8.RuneCountInString(bufBuilder.String()) > app.maxParSize {
				log.Printf("в билдер буфере длинный параграф %v\r\n", utf8.RuneCountInString(bufBuilder.String()))
			}
			b.WriteString(bufBuilder.String())
			bufBuilder.Reset()
		}

		// Если строка не пустая, то производим обработку
		if text != "" {

			//b.WriteString(text)
			// Кол-во символов в склеенных параграфах, предыдущей итерации
			prevLength := utf8.RuneCountInString(b.String())

			// Кол-во символов в текущем обрабатываемом параграфе
			curLength := utf8.RuneCountInString(text)

			// Сумма кол-ва символов в предыдущих склеенных и в текущем параграфах
			concatLength := prevLength + curLength
			//length := utf8.RuneCountInString(b.String()) + utf8.RuneCountInString(text)

			// Если кол-во символов текста в билдере больше максимальной границы maxParSize
			if app.minParSize != 0 && prevLength >= app.maxParSize {
				//log.Printf("cond 1: %v", prevLength)
				app.doSplitting(&b, &bufBuilder)
				//continue
			}

			if app.minParSize != 0 && curLength >= app.maxParSize {
				//log.Println("cond 2")
				bufBuilder.WriteString(text)
				if utf8.RuneCountInString(b.String()) == 0 {
					continue
					//panic(b.String())
				}
			}

			// Если кол-во символов в результирующей строке concatLength менее
			// минимального значения длины параграфа minParSize,
			// то соединяем предыдущие параграфы и текущий обрабатываемый,
			// переходим к следующей итерации цикла и читаем следующий параграф из файла docx,
			// повторяем пока не достигнем границы минимального значения длины параграфа
			if app.minParSize != 0 && concatLength < app.minParSize {
				//log.Println("cond 3")
				b.WriteString(text)
				continue
			}
			// Если кол-во символов в результирующей строке билдера более или равно
			// минимальному значению длины параграфа mixParSize и менее или равно
			// оптимальному значению длины параграфа, то переходим к следующей итерации цикла
			// и читаем следующий параграф из файла docx
			//if app.minParSize != 0 && concatLength >= app.minParSize && concatLength <= app.optParSize {
			if app.minParSize != 0 && concatLength >= app.minParSize && float64(concatLength) <= float64(app.optParSize)*1.05 {
				//log.Println("cond 4")
				// Производим расчет длины текущего параграфа и предыдущих параграфов,
				// набранных до минимального значения границы длины параграфа minParSize
				b.WriteString(text)
				continue
				//cb.WriteString(b.String())
				//} else if curLength < app.optParSize {
			}
			if app.minParSize != 0 && concatLength > app.optParSize && concatLength <= app.maxParSize {
				//log.Printf("cond 5 concatLength: %v", concatLength)
				//log.Printf("cond 5 prevLength: %v", prevLength)
				//log.Printf("cond 5 curLength: %v", curLength)
				bufBuilder.WriteString(text)

				if utf8.RuneCountInString(b.String()) == 0 {
					continue
					//panic(b.String())
				}
			}

			if utf8.RuneCountInString(b.String()) >= app.maxParSize {
				log.Println(utf8.RuneCountInString(b.String()))
				panic(b.String())
			}

			// Записываем текст обработанного параграфа в буфер-билдер
			//bufBuilder.WriteString(text)

			// Если кол-во символов в результирующей строке билдера более
			// максимального значения длины параграфа maxParSize,
			// то вызываем функцию разбивки длинного параграфа DoSplitting
			//if app.minParSize != 0 && length > app.maxParSize {
			//	DoSplitting(b.String())
			//}

			//log.Printf("%v, %v, %v", utf8.RuneCountInString(b.String()) >= app.maxParSize, utf8.RuneCountInString(b.String()), app.maxParSize)
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
		log.Printf("cond 111111111111: %v", utf8.RuneCountInString(b.String()))
		pars = appendParagraph(b, newBook, position, pars)
	}
	b.Reset()

	// Если batchSizeCount меньше batchSize, то записываем оставшиеся параграфы
	if len(pars) > 0 {
		err = app.ps.BulkInsert(ctx, pars, len(pars))
	}

	log.Printf("%v #%v done", newBook.Filename, n+1)
	return nil
}

func (app *App) doSplitting(builder *strings.Builder, bufBuilder *strings.Builder) {
	//log.Printf("Обрабатываем длинный параграф: %v", utf8.RuneCountInString(builder.String()))
	//log.Printf("длина билдер буфера: %v", utf8.RuneCountInString(bufBuilder.String()))

	res := builder.String()
	res = strings.TrimPrefix(res, "<div>")
	res = strings.TrimSuffix(res, "</div>")

	strs := strings.SplitAfter(res, ". ")

	builder.Reset()

	//var splittedPars []string
	var buffer strings.Builder
	var flag bool

	for n, str := range strs {
		//log.Println(str)
		if n == 0 {
			buffer.WriteString("<div>")
		}
		if (utf8.RuneCountInString(buffer.String()) + utf8.RuneCountInString(str) - 4) <= app.optParSize {
			buffer.WriteString(str)
			continue
		}
		if !flag {
			//buffer.WriteString("</div>")
			builder.WriteString(strings.TrimSpace(buffer.String()))
			builder.WriteString("</div>")
			buffer.Reset()
			bufBuilder.WriteString("<div>")
			flag = true
		}

		bufBuilder.WriteString(str)
		//log.Println(builder.String())
		//panic("done1")
	}

	bufBuilder.WriteString("</div>")
	//log.Println(builder.String())
	//log.Println(bufBuilder.String())
	////builder.Reset()
	//panic("done2")
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
