package main

import (
	"encoding/csv"
	"fmt"
	"github.com/audetv/book-parser/parser/docc"
	"github.com/google/uuid"
	flag "github.com/spf13/pflag"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

// Books срез книг
type Books []Book

// Book книга
type Book struct {
	ID   string
	Name string
	Paragraphs
}

// Paragraphs срез параграфов книги
type Paragraphs []Paragraph

// Paragraph параграф из книги
type Paragraph struct {
	ID           uuid.UUID
	RomanNumbers []string
	Text         string
	Position     int
}

// Builder сборщик содержит срез подготовленных параграфов Paragraphs,
// состояние ParagraphsCompleted установлено в true,
// когда сборщик завершил обработку параграфов и начал обработку сносок
// CurrentNote содержит римское число текущей обрабатываемой сноски
// Notes срез подготовленных сносок
type Builder struct {
	Paragraphs          []Paragraph
	ParagraphsCompleted bool
	CurrentNote         string
	Notes               []Note
}

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}

// Note структура для хранения обработанных параграфов,
// в которых римские числа заменены сносками в круглых скобках
type Note struct {
	ParagraphID  string
	RomanNumbers []string
	Paragraph    string
}

type Notes struct {
	CurrentRN string
	Values    []Note
}

// outputPath путь по которому лежат книги для париснга
var outputPath string

func main() {
	flag.StringVarP(
		&outputPath,
		"output",
		"o",
		"./books/VPSSSR/process/",
		"путь хранения файлов для обработки",
	)
	flag.Parse()

	// читаем все файлы в директории
	files, err := os.ReadDir(outputPath)
	if err != nil {
		log.Fatal(err)
	}

	var books Books
	var book Book

	// итерируемся по списку файлов
	for _, file := range files {
		fmt.Println(file.Name(), file.IsDir())

		if file.IsDir() == false {
			// если файл .gitignore, то ничего не делаем пропускаем и продолжаем цикл
			if file.Name() == ".gitignore" {
				continue
			}
			bookName := file.Name()

			// Генерируем UUID
			ID := uuid.New().String()

			// создаём книгу и срез параграфов в ней
			paragraphs := Paragraphs{}
			book = Book{ID, bookName, paragraphs}

			book = parseParagraphs(book, file)
			// добавляем созданную книгу в срез книг
			books = append(books, book)
		}
	}
	currentTime := time.Now()
	// записываем книги в файлы csv vertex_book.csv, vertex_paragraph.csv, edge_belongTo.csv, edge_follow.csv
	writeBook(books, fmt.Sprintf("./csv/%v_book.csv", currentTime.Format("150405_02012006")))
}

func parseParagraphs(book Book, file os.DirEntry) Book {
	fp := filepath.Clean(fmt.Sprintf("%v%v", outputPath, file.Name()))
	r, err := docc.NewReader(fp)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	builder := new(Builder)
	// position номер параграфа в индексе
	position := 1

	// Список параграфов и сносок к ним
	//var notes Notes

	for {
		p, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		// Если строка не пустая, то записываем в индекс
		if p != "" {

			//ID := uuid.New().String()
			//paragraph := Paragraph{ID, p, strconv.Itoa(position)}

			// если эту строку пропустить, удалить, то не нужны будут функции replaceParagraph?
			//book.Paragraphs = append(book.Paragraphs, paragraph)
			//builder.Paragraphs = append(book.Paragraphs, paragraph)

			// формируем и получаем список сносок

			// это условие надо положить внутрь функции builder.processParagraph и ее переименовать
			if builder.ParagraphsCompleted == false {

				builder.processParagraph(p, position)

			} else {
				log.Printf("Секция сносок position: %v\r\n", position)
				//notes = builder.processParagraphNote(p)
			}

			position++
		}
	}

	log.Println(builder.Paragraphs)

	// обработка подготовленных параграфов со вставленными сносками в круглых скобках
	//for n, paragraph := range book.Paragraphs {
	//	book.Paragraphs[n] = replaceParagraph(paragraph, notes)
	//}

	return book
}

//func replaceParagraph(paragraph Paragraph, notes Notes) Paragraph {
//	for _, note := range notes.Values {
//		if paragraph.ID == note.ParagraphID {
//			paragraph.Text = note.Paragraph
//		}
//	}
//	return paragraph
//}

// processParagraphNote функция проверяет строку-параграф на наличие в ней римского числа,
// заключенного в квадратные скобки, формирует срез сносок и производит замену
// римских чисел в обработанных параграфах на сноски заключенные в круглые скобки
func (b *Builder) processParagraph(p string, position int) {

	matched := regexp.MustCompile(`\[M{0,3}(CM|CD|D?C{0,3})?(XC|XL|L?X{0,3})?(IX|IV|V?I{0,3})?]`)

	// возвращает срез совпадений римского числа в строке romanian number, может быть более одной сноски в сроке
	rns := matched.FindAllString(p, -1)

	// Если римское число найдено, то записываем обрабатываем его
	//for _, rn := range rns {
	// Итерируемся по срезу записанных сносок, для того чтобы найти уже сохраненные числа, сноски
	for _, paragraph := range b.Paragraphs {

		// Если в параграфе нет ни одной сноски, len(rns) == 0, и если установлено значение notes.CurrentRN
		// То надо этот параграф добавить к предыдущему параграфу в сноску, как?

		for _, rn := range rns {
			// Если римское число в записанной сноске равно числу найденному в переданном для обработки параграфе
			for _, noteRN := range paragraph.RomanNumbers {
				if noteRN == rn {
					b.ParagraphsCompleted = true

					//// записываем в Notes, что обрабатываем сноску с номером rn
					//log.Printf("обрабатываем сноску с номером: %v \r\n", rn)
					////log.Println(i, note.RomanNumbers[k], rn)
					//// сначала удаляем из текущего параграфа сноски римское число,
					//// а теги параграфа заменяем на круглые скобки
					//replacer := strings.NewReplacer(rn, "", "<p>", " (", "</p>", ")")
					//newP := replacer.Replace(strings.TrimSpace(p))
					//
					//// после заменяем римское число на подготовленную сноску, заключенную в круглые скобки
					//replacer = strings.NewReplacer(rn, newP)
					//paragraph.Text = replacer.Replace(paragraph.Text)
					//
					//// заменяем старую сноску, обработанной сноской
					//b.Paragraphs[n] = paragraph
					//state = true
				}
			}
		}
	}

	// создаём объект сноски и записываем в него срез римских чисел
	if b.ParagraphsCompleted == false {
		paragraph := Paragraph{
			ID:           uuid.New(),
			RomanNumbers: rns,
			Text:         p,
			Position:     position,
		}

		b.Paragraphs = append(b.Paragraphs, paragraph)
	}
}

func writeBook(books Books, outputPath string) {
	// Data array to write to CSV
	var data [][]string
	var dataParagraphs [][]string

	for _, book := range books {
		data = append(data, []string{
			// Make sure the property order here matches
			// the one from 'headerRow' !!!
			book.ID,
			book.Name,
		})

		for _, paragraph := range book.Paragraphs {
			dataParagraphs = append(dataParagraphs, []string{
				paragraph.ID.String(),
				paragraph.Text,
				strconv.Itoa(paragraph.Position),
			})
		}
	}

	writeCSV(data, outputPath)
	currentTime := time.Now()
	writeCSV(dataParagraphs, fmt.Sprintf("./csv/%v_paragraph.csv", currentTime.Format("150405_02012006")))
}

func writeCSV(data [][]string, path string) {
	// Create file
	file, err := os.Create(path)
	checkError("Cannot create file", err)
	defer file.Close()

	// Create writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write rows into file
	for _, value := range data {
		err = writer.Write(value)
		checkError("Cannot write to file", err)
	}
}
