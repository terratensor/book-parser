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
	"strings"
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
	ID       string
	Text     string
	Position string
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

	// position номер параграфа в индексе
	position := 1

	// Список параграфов и сносок к ним
	var notes Notes

	for {
		p, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		// Если строка не пустая, то записываем в индекс
		if p != "" {

			ID := uuid.New().String()
			paragraph := Paragraph{ID, p, strconv.Itoa(position)}

			// если эту строку пропустить, удалить, то не нужны будут функции replaceParagraph?
			book.Paragraphs = append(book.Paragraphs, paragraph)
			// формируем и получаем список сносок
			notes = processParagraphNote(paragraph, notes)

			position++
		}
	}

	log.Println(notes)

	// обработка подготовленных параграфов со вставленными сносками в круглых скобках
	for n, paragraph := range book.Paragraphs {
		book.Paragraphs[n] = replaceParagraph(paragraph, notes)
	}

	return book
}

func replaceParagraph(paragraph Paragraph, notes Notes) Paragraph {
	for _, note := range notes.Values {
		if paragraph.ID == note.ParagraphID {
			paragraph.Text = note.Paragraph
		}
	}
	return paragraph
}

// processParagraphNote функция проверяет строку-параграф на наличие в ней римского числа,
// заключенного в квадратные скобки, формирует срез сносок и производит замену
// римских чисел в обработанных параграфах на сноски заключенные в круглые скобки
func processParagraphNote(p Paragraph, notes Notes) Notes {

	matched := regexp.MustCompile(`\[M{0,3}(CM|CD|D?C{0,3})?(XC|XL|L?X{0,3})?(IX|IV|V?I{0,3})?]`)

	// возвращает срез совпадений римского числа в строке romanian number, может быть более одной сноски в сроке
	rns := matched.FindAllString(p.Text, -1)

	state := false
	// Если римское число найдено, то записываем обрабатываем его
	//for _, rn := range rns {
	// Итерируемся по срезу записанных сносок, для того чтобы найти уже сохраненные числа, сноски
	for n, note := range notes.Values {
		for _, rn := range rns {
			// Если римское число в записанной сноске равно числу найденному в переданном для обработки параграфе
			for _, noteRN := range note.RomanNumbers {
				if noteRN == rn {
					//log.Println(i, note.RomanNumbers[k], rn)
					// сначала удаляем из текущего параграфа сноски римское число,
					// а теги параграфа заменяем на круглые скобки
					replacer := strings.NewReplacer(rn, "", "<p>", " (", "</p>", ")")
					newP := replacer.Replace(strings.TrimSpace(p.Text))

					// после заменяем римское число на подготовленную сноску, заключенную в круглые скобки
					replacer = strings.NewReplacer(rn, newP)
					note.Paragraph = replacer.Replace(note.Paragraph)

					// заменяем старую сноску, обработанной сноской
					notes.Values[n] = note
					state = true
				}
			}
		}
	}

	// создаём объект сноски и записываем в него срез римских чисел
	if state {
		state = false
	} else {
		newNote := Note{
			ParagraphID:  p.ID,
			RomanNumbers: rns,
			Paragraph:    p.Text,
		}
		notes.Values = append(notes.Values, newNote)
	}

	return notes
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
				paragraph.ID,
				paragraph.Text,
				paragraph.Position,
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
