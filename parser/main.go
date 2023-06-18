package main

import (
	"fmt"
	"github.com/audetv/book-parser/parser/docc"
	"github.com/manticoresoftware/go-sdk/manticore"
	flag "github.com/spf13/pflag"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Paragraphs срез параграфов книги
type Paragraphs []Paragraph

// Paragraph параграф из книги
type Paragraph struct {
	Text     string
	Position int
	Book     string
}

var wg sync.WaitGroup

var outputPath string

func main() {

	//ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	flag.StringVarP(&outputPath, "output", "o", "./books/VPSSSR/process/", "путь хранения файлов для обработки")
	flag.Parse()

	cl := manticore.NewClient()
	cl.SetServer("localhost", 9312)
	_, err := cl.Open()
	if err != nil {
		fmt.Printf("Conn: %v", err)
	}

	// удаляем таблицу
	dropTable(cl)
	// создаём таблицу
	createTable(cl)
	// читаем все файлы в директории
	files, err := os.ReadDir(outputPath)
	if err != nil {
		log.Fatal(err)
	}

	c := make(chan Paragraphs)
	log.Printf("файлов в обработке: %v", len(files))

	for _, file := range files {
		//fmt.Println(file.Name(), file.IsDir())

		wg.Add(1)

		if file.IsDir() == false {
			if file.Name() == ".gitignore" {
				return
			}
			bookName := file.Name()

			go createIndex(c, bookName, file)
		}

	}

	for i := 0; i < len(files); i++ {
		paragraphs := <-c
		createBulkRecord(cl, paragraphs)

		log.Printf("файл №%v, параграфы сохранены в мантикоре!", i+1)
	}

	wg.Wait()
	fmt.Println("Все файлы обработаны")

	//bookName := "Время - начинаю про Сталина рассказ….docx"
	// bookName := "Об имитационно-провокационной деятельности"
	// bookName := "“Мастер и Маргарита”:\nгимн демонизму? \nлибо\nЕвангелие беззаветной веры\n"
	//createIndex(cl, bookName)
	// truncateTable(cl)
	//createTable(cl)
	// dropTable(cl)

	// fp := filepath.Clean("./books/dotu.docx")
	// r, err := docc.NewReader(fp)
	// if err != nil {
	// 	panic(err)
	// }
	// defer r.Close()
	//
	// for i := 0; ; i++ {
	// 	p, err := r.Read()
	// 	if err == io.EOF {
	// 		return
	// 	} else if err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Printf("%d  %v\r\n", i, p)
	// 	// do something with p:string
	// }
}

func createIndex(c chan Paragraphs, bookName string, file os.DirEntry) {

	defer wg.Done()

	var paragraph Paragraph
	var paragraphs Paragraphs

	fp := filepath.Clean(fmt.Sprintf("%v%v", outputPath, file.Name()))
	r, err := docc.NewReader(fp)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	// position номер параграфа в индексе
	position := 1
	for {
		p, err := r.Read()
		if err == io.EOF {
			log.Printf("%v параграфов обработано: %v", bookName, position)
			c <- paragraphs
		} else if err != nil {
			panic(err)
		}

		// Если строка не пустая, то записываем в индекс
		if p != "" {
			paragraph = Paragraph{Text: p, Position: position, Book: bookName}
			paragraphs = append(paragraphs, paragraph)
			//createRecord(cl, p, position, bookName)
			//log.Printf("%d  %v\r\n", position, p)
			position++
		}
	}

}

func createRecord(cl manticore.Client, p string, i int, bookName string) {

	// fmt.Printf("%d %v", n, comment)
	// q := "INSERT INTO ads_search (title) VALUES (\"" + ad.Title + "\")"
	// str := fmt.Sprintf("replace into viewquestion values(%d, '%s','%s', '%s', '%s', '%s')",
	escapedParagraph := EscapeString(p)
	str := fmt.Sprintf(`insert into booksearch(text,position,type,book,datetime) values('%v', '%v', '%v', '%v', '%v')`, escapedParagraph, i, 1, bookName, time.Now().Unix())
	res, _ := cl.Sphinxql(str)
	// cl.Sphinxql(str)
	//
	// fmt.Printf("%v\r\n", res[0].Msg)
	if strings.Contains(fmt.Sprintf("%v", res), "ERROR") {
		fmt.Println(p)
		fmt.Println(res)
		//panic(res)
	}
	// fmt.Println(res)
}

func createBulkRecord(cl manticore.Client, paragraphs Paragraphs) {

	defer duration(track("сохранено в мантикору за"))

	var values string
	end := ","

	for n, p := range paragraphs {
		escapedParagraph := EscapeString(p.Text)
		if len(paragraphs) == n+1 {
			end = ";"
			log.Printf(p.Book)
		}
		values += fmt.Sprintf("('%v', '%v', '%v', '%v', '%v')%v ", escapedParagraph, p.Position, 1, p.Book, time.Now().Unix(), end)
	}

	str := fmt.Sprintf(`INSERT INTO booksearch(text,position,type,book,datetime) VALUES %v`, values)
	res, _ := cl.Sphinxql(str)

	if strings.Contains(fmt.Sprintf("%v", res), "ERROR") {
		fmt.Println(res)
	}
}

func createTable(cl manticore.Client) {
	res, err := cl.Sphinxql("create table booksearch(`text` text, position integer, type integer, book string, datetime timestamp) morphology='stem_ru'")
	fmt.Println(res, err)
}

func dropTable(cl manticore.Client) {
	res, err := cl.Sphinxql(`drop table booksearch`)
	//fmt.Println(res, err)
	log.Println(res, err)
}

// EscapeString escapes characters that are treated as special operators by the query language parser.
//
// `from` is a string to escape.
// Returns escaped string.
func EscapeString(from string) string {
	dest := make([]byte, 0, 2*len(from))
	for i := 0; i < len(from); i++ {
		c := from[i]
		switch c {
		case '\\', '(', ')', '|', '-', '!', '@', '~', '"', '\'', '&', '/', '^', '$', '=', '<':
			dest = append(dest, '\\')
		}
		dest = append(dest, c)
	}
	return string(dest)
}

func track(msg string) (string, time.Time) {
	return msg, time.Now()
}

func duration(msg string, start time.Time) {
	log.Printf("%v: %v\n", msg, time.Since(start))
}
