package main

import (
	"fmt"
	"github.com/audetv/book-parser/parser/docc"
	"github.com/manticoresoftware/go-sdk/manticore"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {

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
	files, err := os.ReadDir("books/VPSSSR/DOCX")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		//fmt.Println(file.Name(), file.IsDir())

		if file.IsDir() == false {
			bookName := file.Name()
			createIndex(cl, bookName, file)
		}

	}

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

func createIndex(cl manticore.Client, bookName string, file os.DirEntry) {
	// fp := filepath.Clean("./books/master-i-margarita.docx")
	// fp := filepath.Clean("./books/ob_imitac-prov_deyat_a4-20010324.docx")
	fp := filepath.Clean(fmt.Sprintf("./books/VPSSSR/DOCX/%v", file.Name()))
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
			return
		} else if err != nil {
			panic(err)
		}

		// Если строка не пустая, то записываем в индекс
		if p != "" {
			log.Printf("%d  %v\r\n", position, p)
			createRecord(cl, p, position, bookName)
			position++
		}
	}

}

func createRecord(cl manticore.Client, p string, i int, bookName string) {

	// fmt.Printf("%d %v", n, comment)
	// q := "INSERT INTO ads_search (title) VALUES (\"" + ad.Title + "\")"
	// str := fmt.Sprintf("replace into viewquestion values(%d, '%s','%s', '%s', '%s', '%s')",
	str := fmt.Sprintf(`insert into booksearch( 
                       text,
                       position,
                       type,
                       book,
                       datetime
                       ) values('%v','%v', 1, '%v', 0 )`, p, i, bookName)
	res, _ := cl.Sphinxql(str)
	// cl.Sphinxql(str)
	//
	// fmt.Printf("%v\r\n", res[0].Msg)
	if strings.Contains(fmt.Sprintf("%v", res), "ERROR") {
		fmt.Println(res)
	}
	// fmt.Println(res)
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
