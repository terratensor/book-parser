package main

import (
	"encoding/csv"
	"fmt"
	"github.com/audetv/book-parser/parser/docc"
	"github.com/bwmarrin/snowflake"
	flag "github.com/spf13/pflag"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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

// Belongs edge срез ребер графа, принадлежность параграфа к книге
type Belongs []BelongToBook

// BelongToBook параграф ID принадлежит книге ID
type BelongToBook struct {
	BookID      string
	ParagraphID string
	Rank        string
}

// Follows edge срез ребер графа, последовательность параграфов
type Follows []FollowParagraph

// FollowParagraph последовательность параграфов, следующий параграф
type FollowParagraph struct {
	ParagraphID     string
	NextParagraphID string
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

// outputPath путь по которому лежат книги для париснга
var outputPath string

func main() {
	flag.StringVarP(&outputPath, "output", "o", "./books/VPSSSR/process/", "путь хранения файлов для обработки")
	flag.Parse()

	// Создаём новый узел Node с номером 1 для генерации IDs по алгоритму snowflake
	node, err := snowflake.NewNode(1)
	if err != nil {
		fmt.Println(err)
		return
	}

	// читаем все файлы в директории
	files, err := os.ReadDir(outputPath)
	if err != nil {
		log.Fatal(err)
	}

	var books Books
	var book Book
	var belongs Belongs
	var follows Follows

	// итерируемся по списку файлов
	for _, file := range files {
		fmt.Println(file.Name(), file.IsDir())

		if file.IsDir() == false {
			// если файл .gitignore, то ничего не делаем пропускаем и продолжаем цикл
			if file.Name() == ".gitignore" {
				continue
			}
			bookName := file.Name()

			// создаём ID 32 разрядный md5 из наименования файла книги
			//data := []byte(bookName)
			//id := fmt.Sprintf("%x", md5.Sum(data))

			// Генерируем числовой ИД по алгоритму snowflake.
			ID := node.Generate().String()

			// создаём книгу и срез параграфов в ней
			paragraphs := Paragraphs{}
			book = Book{ID, bookName, paragraphs}

			book, belongs, follows = parseParagraphs(node, book, belongs, follows, file)
			// добавляем созданную книгу в срез книг
			books = append(books, book)
		}
	}
	// записываем книги в файлы csv vertex_book.csv, vertex_paragraph.csv, edge_belongTo.csv, edge_follow.csv
	writeBook(books, "./csv/vertex_book.csv")
	writeBelongs(belongs, "./csv/edge_belongs.csv")
	writeFollows(follows, "./csv/edge_follows.csv")
}

func parseParagraphs(node *snowflake.Node, book Book, belongs Belongs, follows Follows, file os.DirEntry) (Book, Belongs, Follows) {
	fp := filepath.Clean(fmt.Sprintf("%v%v", outputPath, file.Name()))
	r, err := docc.NewReader(fp)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	// position номер параграфа в индексе
	position := 1

	// Список параграфов и сносок к ним
	var notes []Note

	for {
		p, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		// Если строка не пустая, то записываем в индекс
		if p != "" {
			// создаём ID 32 разрядный md5 из наименования файла книги
			//data := []byte(p)
			//id := fmt.Sprintf("%x", md5.Sum(data))
			ID := node.Generate().String()

			paragraph := Paragraph{ID, p, strconv.Itoa(position)}

			book.Paragraphs = append(book.Paragraphs, paragraph)
			// формируем и получаем список сносок
			notes = processParagraphNote(paragraph, notes)

			belongs = append(belongs, BelongToBook{book.ID, paragraph.ID, strconv.Itoa(position)})

			// Если это не первый элемент среза
			if len(follows) > 0 {
				// записываем последний элемент среза в переменную prev, в ней содержится ID предыдущего параграфа,
				// к которому надо добавить ID текущего параграфа
				prev := follows[len(follows)-1]
				// удаляем последний элемент среза
				follows = follows[:len(follows)-1]

				// Восстанавливаем последний элемент среза, записываем в него ID предыдущего параграфа и ID текущего параграфа
				follows = append(follows, FollowParagraph{prev.ParagraphID, paragraph.ID})
			}

			// Подготавливаем следующий элемент среза, записываем в него ID текущего параграфа и пустую строку
			// так как ID следующего параграфа нам еще не известно
			follows = append(follows, FollowParagraph{paragraph.ID, "0"})

			position++
		}
	}

	log.Println(notes)

	// обработка подготовленных параграфов со вставленными сносками в круглых скобках
	for n, paragraph := range book.Paragraphs {
		book.Paragraphs[n] = replaceParagraph(paragraph, notes)
	}

	return book, belongs, follows
}

func replaceParagraph(paragraph Paragraph, notes []Note) Paragraph {
	for _, note := range notes {
		if paragraph.ID == note.ParagraphID {
			paragraph.Text = note.Paragraph
		}
	}
	return paragraph
}

// processParagraphNote функция проверяет строку-параграф на наличие в ней римского числа,
// заключенного в квадратные скобки, формирует срез сносок и производит замену
// римских чисел в обработанных параграфах на сноски заключенные в круглые скобки
func processParagraphNote(p Paragraph, notes []Note) []Note {

	matched := regexp.MustCompile(`\[M{0,3}(CM|CD|D?C{0,3})?(XC|XL|L?X{0,3})?(IX|IV|V?I{0,3})?]`)

	// возвращает срез совпадений римского числа в строке romanian number, может быть более одной сноски в сроке
	rns := matched.FindAllString(p.Text, -1)

	//if len(notes) == 0 {
	//	newNote := Note{
	//		ParagraphID:  p.ID,
	//		RomanNumbers: rns,
	//		Paragraph:    p.Text,
	//	}
	//	notes = append(notes, newNote)
	//} else {
	//	for _, rn := range rns {
	//		if rn == "[I]" {
	//			return notes, true
	//		}
	//	}
	//}
	//
	//return notes, false

	//log.Printf(" stack roman number: %v\n", n)

	state := false
	// Если римское число найдено, то записываем обрабатываем его
	//for _, rn := range rns {
	// Итерируемся по срезу записанных сносок, для того чтобы найти уже сохраненные числа, сноски
	for n, note := range notes {
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
					notes[n] = note
					state = true
				}
			}
		}
	}

	//if state {
	//	state = false
	//} else {
	//	newNote := Note{
	//		ParagraphID: p.ID,
	//		RomanNumber: rn,
	//		Paragraph:   p.Text,
	//	}
	//	notes = append(notes, newNote)
	//}

	//}

	// создаём объект сноски и записываем в него срез римских чисел
	if state {
		state = false
	} else {
		newNote := Note{
			ParagraphID:  p.ID,
			RomanNumbers: rns,
			Paragraph:    p.Text,
		}
		notes = append(notes, newNote)
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
	writeCSV(dataParagraphs, "./csv/vertex_paragraph.csv")
}

func writeBelongs(belongs Belongs, outputPath string) {
	var data [][]string

	for _, belong := range belongs {
		data = append(data, []string{
			// Make sure the property order here matches
			// the one from 'headerRow' !!!
			belong.BookID,
			belong.ParagraphID,
			belong.Rank,
		})
	}

	writeCSV(data, outputPath)
}

func writeFollows(follows Follows, outputPath string) {
	var data [][]string

	for _, follow := range follows {
		data = append(data, []string{
			// Make sure the property order here matches
			// the one from 'headerRow' !!!
			follow.ParagraphID,
			follow.NextParagraphID,
		})
	}

	writeCSV(data, outputPath)
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
