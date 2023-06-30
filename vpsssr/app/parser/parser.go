package parser

import (
	"encoding/csv"
	"fmt"
	"github.com/audetv/book-parser/parser/docc"
	"github.com/google/uuid"
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
	ID       uuid.UUID
	Name     string
	Filename string
	Paragraphs
}

// Paragraphs срез параграфов книги
type Paragraphs []Paragraph

// Paragraph параграф из книги
type Paragraph struct {
	ID           uuid.UUID
	RomanNumbers []string // Срез римских чисел содержащихся в параграфе
	Text         string
	Position     int
	Length       int
}

// Builder сборщик содержит срез подготовленных параграфов Paragraphs,
// состояние ParagraphsCompleted установлено в true,
// когда сборщик завершил обработку параграфов и начал обработку сносок
// CurrentNote содержит римское число текущей обрабатываемой сноски
// Notes map подготовленных сносок
type Builder struct {
	Paragraphs          []Paragraph
	ParagraphsCompleted bool
	CurrentNote         string
	Notes               *Notes
}

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}

// Notes структура для хранения обработанных сносок,
type Notes struct {
	m map[string]string
}

// outputPath путь по которому лежат книги для париснга
var outputPath string

func NewNotes() *Notes {
	return &Notes{
		m: make(map[string]string),
	}
}

func Parse(path string) Books {
	// читаем все файлы в директории
	files, err := os.ReadDir(path)
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

			var filename = file.Name()
			var extension = filepath.Ext(filename)
			var bookName = filename[0 : len(filename)-len(extension)]

			// Генерируем UUID
			ID := uuid.New()

			// создаём книгу и срез параграфов в ней
			paragraphs := Paragraphs{}
			book = Book{ID, bookName, filename, paragraphs}

			book = parseParagraphs(book, file, path)
			// добавляем созданную книгу в срез книг
			books = append(books, book)
		}
	}

	return books
	// записываем книги в БД
	//writeBooks(ctx, app, books)
}

func parseParagraphs(book Book, file os.DirEntry, outputPath string) Book {
	fp := filepath.Clean(fmt.Sprintf("%v%v", outputPath, file.Name()))
	log.Println(fp)
	r, err := docc.NewReader(fp)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	builder := new(Builder)
	builder.Notes = NewNotes()
	// position номер параграфа в индексе
	position := 1

	for {
		p, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		// Если строка не пустая, то записываем в индекс
		if p != "" {

			builder.processParagraph(p, position)
			position++
		}
	}

	builder.mergeNotes()
	book.Paragraphs = builder.Paragraphs

	return book
}

// processParagraph функция формирует строку-параграф, проверяет строку на наличие в ней римского числа,
// заключенного в квадратные скобки, после обработки всех параграфов формирует срез сносок
func (b *Builder) processParagraph(p string, position int) {

	matched := regexp.MustCompile(`\[M{0,3}(CM|CD|D?C{0,3})?(XC|XL|L?X{0,3})?(IX|IV|V?I{0,3})?]`)

	// возвращает срез совпадений римского числа в строке romanian number, может быть более одной сноски в сроке
	rns := matched.FindAllString(p, -1)

	for _, paragraph := range b.Paragraphs {

		for _, rn := range rns {
			// Если римское число в записанной сноске равно числу найденному в переданном для обработки параграфе,
			// то это означает, что началась обработка сносок и параграфы закончились
			for _, noteRN := range paragraph.RomanNumbers {
				if noteRN == rn {
					b.ParagraphsCompleted = true
					// сохраняем текущее римское число сноски, для склейки сносок в одну
					b.CurrentNote = rn
				}
			}
		}
	}

	// создаём параграф и записываем в него срез римских чисел
	// сносок, которые содержит параграф
	if b.ParagraphsCompleted == false {
		paragraph := Paragraph{
			ID:           uuid.New(),
			RomanNumbers: rns,
			Text:         p,
			Position:     position,
		}

		b.Paragraphs = append(b.Paragraphs, paragraph)
	} else {
		// Заменяем римское число в квадратных скобках на пустую строку, тэг <p> на <span>
		// семантически неверно внутри тега p параграфа, помещать вложенные параграфы,
		// поэтому меняем тег параграфа сноски на тег span
		replacer := strings.NewReplacer(b.CurrentNote, "", "<div>", "<p>", "</div>", "</p>")
		noteText := replacer.Replace(strings.TrimSpace(p))
		// Соединяет строки сноски, все бывшие параграфы, теперь span в одну строку
		result := strings.Join([]string{b.Notes.m[b.CurrentNote], noteText}, "")
		b.Notes.m[b.CurrentNote] = result
	}
}

func (b *Builder) mergeNotes() {
	for n, paragraph := range b.Paragraphs {
		for _, rn := range paragraph.RomanNumbers {
			note := b.Notes.m[rn]

			// Заменяем римское число на подготовленную сноску, заключенную в круглые скобки
			replacer := strings.NewReplacer(rn, fmt.Sprintf("(%v)", note))
			paragraph.Text = replacer.Replace(paragraph.Text)

			// заменяем старый параграф, обработанным параграфом со вставленной в него сноской
			b.Paragraphs[n] = paragraph
		}
	}
}

func writeBook(books Books, outputPath string) {
	headerRow := []string{
		"uuid", "filename", "name",
	}
	headerRowParagraph := []string{
		"uuid", "book_uuid", "text", "position",
	}
	// Data array to write to CSV
	data := [][]string{
		headerRow,
	}
	dataParagraphs := [][]string{
		headerRowParagraph,
	}

	for _, book := range books {
		data = append(data, []string{
			// Make sure the property order here matches
			// the one from 'headerRow' !!!
			book.ID.String(),
			book.Name,
			book.Name,
		})

		for _, paragraph := range book.Paragraphs {
			dataParagraphs = append(dataParagraphs, []string{
				paragraph.ID.String(),
				book.ID.String(),
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
