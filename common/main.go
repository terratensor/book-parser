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
	"strconv"
	"time"
)

// Books срез книг
type Books []Book

// Book книга
type Book struct {
	ID   uuid.UUID
	Name string
	Paragraphs
}

// Paragraphs срез параграфов книги
type Paragraphs []Paragraph

// Paragraph параграф из книги
type Paragraph struct {
	ID       uuid.UUID
	Text     string
	Position int
}

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
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
			ID := uuid.New()

			// создаём книгу и срез параграфов в ней
			paragraphs := Paragraphs{}
			book = Book{ID, bookName, paragraphs}

			book.parse(file)
			book.save()

			// добавляем созданную книгу в срез книг
			books = append(books, book)
		}
	}

	// записываем книги в файлы csv vertex_book.csv, vertex_paragraph.csv, edge_belongTo.csv, edge_follow.csv
	books.save()
}

func (b *Book) parse(file os.DirEntry) {
	fp := filepath.Clean(fmt.Sprintf("%v%v", outputPath, file.Name()))
	r, err := docc.NewReader(fp)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	// position номер параграфа в индексе
	position := 1

	for {
		text, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		// Если строка не пустая, то записываем в индекс
		if text != "" {
			paragraph := Paragraph{
				ID:       uuid.New(),
				Text:     text,
				Position: position,
			}

			b.Paragraphs = append(b.Paragraphs, paragraph)
			position++
		}
	}
}

func (b *Book) save() {
	headerRowBook := []string{
		"uuid", "book_uuid", "paragraph", "position",
	}

	data := [][]string{
		headerRowBook,
	}

	for _, paragraph := range b.Paragraphs {
		data = append(data, []string{
			paragraph.ID.String(),
			b.ID.String(),
			paragraph.Text,
			strconv.Itoa(paragraph.Position),
		})
	}

	currentTime := time.Now()
	writeCSV(data, fmt.Sprintf("./csv/%v_paragraph-%v.csv", b.Name, currentTime.Format("150405_02012006")))
}

func (bs *Books) save() {
	headerRow := []string{
		"uuid", "filename", "name",
	}

	// Data array to write to CSV
	data := [][]string{
		headerRow,
	}

	for _, book := range *bs {
		data = append(data, []string{
			// Make sure the property order here matches
			// the one from 'headerRow' !!!
			book.ID.String(),
			book.Name,
			book.Name,
		})
	}

	currentTime := time.Now()
	writeCSV(data, fmt.Sprintf("./csv/%v_books.csv", currentTime.Format("150405_02012006")))
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
