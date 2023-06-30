package docc

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var ErrNotSupportFormat = errors.New("the file is not supported")

type Document struct {
	XMLName xml.Name `xml:"document"`
	Body    struct {
		P []struct {
			R []struct {
				T struct {
					Text  string `xml:",chardata"`
					Space string `xml:"space,attr"`
				} `xml:"t"`
			} `xml:"r"`
		} `xml:"p"`
	} `xml:"body"`
}

type Paragraph struct {
	R []struct {
		T struct {
			Text  string `xml:",chardata"`
			Space string `xml:"space,attr"`
		} `xml:"t"`
	} `xml:"r"`
}

type Reader struct {
	docxPath string
	fromDoc  bool
	docx     *zip.ReadCloser
	xml      io.ReadCloser
	dec      *xml.Decoder
}

type FootnoteReference struct {
	id string
}

// NewReader создаёт Reader структуру.
// После прочтения, структура Reader должна быть закрыта Close().
func NewReader(docxPath string) (*Reader, error) {
	r := new(Reader)
	r.docxPath = docxPath
	ext := strings.ToLower(filepath.Ext(docxPath))
	if ext != ".docx" {
		return nil, ErrNotSupportFormat
	}

	a, err := zip.OpenReader(r.docxPath)
	if err != nil {
		return nil, err
	}
	r.docx = a

	f, err := a.Open("word/document.xml")
	if err != nil {
		return nil, err
	}
	r.xml = f
	r.dec = xml.NewDecoder(f)

	return r, nil
}

// Read читает файл .docx по параграфам.
// Если параграфы в файле закончились, возвращает ошибку io.EOF.
func (r *Reader) Read() (string, error) {
	err := seekNextTag(r.dec, "p")
	if err != nil {
		return "", err
	}
	p, err := seekParagraph(r.dec)
	if err != nil {
		return "", err
	}
	return p, nil
}

// ReadAll считывает весь файл .docx целиком. Возвращает срез параграфов и ошибку.
func (r *Reader) ReadAll() ([]string, error) {
	var ps []string
	for {
		// p - прочитанный параграф
		p, err := r.Read()
		// Если вернулась ошибка io.EOF значит параграфы в файле закончились,
		// возвращает заполненный срез, если вернулась другая ошибка, то возвращает ошибку
		if err == io.EOF {
			return ps, nil
		} else if err != nil {
			return nil, err
		}
		// ps - срез параграфов
		ps = append(ps, p)
	}
}

func (r *Reader) Close() error {
	r.xml.Close()
	r.docx.Close()
	if r.fromDoc {
		os.Remove(r.docxPath)
	}
	return nil
}

func seekParagraph(dec *xml.Decoder) (string, error) {
	var t string
	var tag, headerTag string
	var fr FootnoteReference
	for {
		token, err := dec.Token()
		if err != nil {
			return "", err
		}
		switch tt := token.(type) {
		case xml.EndElement:
			if tt.Name.Local == "p" {
				// вырезаем мусор
				t = cutOutTrash(t)
				// Удаляет лишние пробелы в начал и в конце строки
				t = strings.TrimSpace(t)
				// Если строка пустая, то возвращаем строку и ничего не даем
				if t == "" {
					return t, nil
				}
				// Обрамляем строку нужным html тегом
				t = wrapperHtmlTag(headerTag, t)
				headerTag = ""
				return t, nil
			}
		case xml.StartElement:
			if tt.Name.Local == "pStyle" {
				for _, attr := range tt.Attr {
					headerTag = getHeaderTag(attr.Value)
				}
			}
			//if tt.Name.Local == "b" {
			//	tag = "b"
			//}
			//if tt.Name.Local == "i" {
			//	tag = "i"
			//}
			// Ищем ссылку на footnoteReference и присваевает номер сноски в id
			if tt.Name.Local == "footnoteReference" {
				fr.id = tt.Attr[0].Value
			}
			if tt.Name.Local == "t" {
				text, err := seekText(dec)
				if err != nil {
					return "", err
				}

				switch tag {
				case "b":
					t = t + fmt.Sprintf("<b>%v</b>", text)
					tag = ""
				case "i":
					t = t + fmt.Sprintf("<i>%v</i>", text)
					tag = ""
				default:
					t = t + text
				}
			}
		}
	}
}

// cutOutTrash определяет мусорные строки по регулярным выражениям
// и возвращает пустую строку или строку без изменений
func cutOutTrash(t string) string {
	// Проверяет, что строка не является строкой типа «*         *         *»
	// Проверяет, что строка не является строкой типа «—————————»
	// Проверяет, что строка не является строкой типа «•••••••»
	// Проверяет, что строка не является строкой типа «▬▬▬▬▬▬▬»
	// Проверяет, что строка не является строкой типа «______»
	var validT = regexp.MustCompile(`^[\s*—•▬_]+$`)
	if validT.MatchString(t) == true {
		t = ""
	}
	return t
}

func wrapperHtmlTag(headerTag string, t string) string {
	switch headerTag {
	case "h1":
		t = fmt.Sprintf("<h1>%v</h1>", t)
		headerTag = ""
	case "h2":
		t = fmt.Sprintf("<h2>%v</h2>", t)
		headerTag = ""
	case "h3":
		t = fmt.Sprintf("<h3>%v</h3>", t)
		headerTag = ""
	case "h4":
		t = fmt.Sprintf("<h4>%v</h4>", t)
		headerTag = ""
	case "h5":
		t = fmt.Sprintf("<h5>%v</h5>", t)
		headerTag = ""
	case "h6":
		t = fmt.Sprintf("<h6>%v</h6>", t)
		headerTag = ""
	default:
		t = fmt.Sprintf("<div>%v</div>", t)
	}
	return t
}

// getHeaderTag возвращает html tag заголовка
func getHeaderTag(value string) string {
	var tag string
	switch value {
	case "Heading1":
		tag = "h1"
	case "Heading2":
		tag = "h2"
	case "Heading3":
		tag = "h3"
	case "Heading4":
		tag = "h4"
	case "Heading5":
		tag = "h5"
	case "Heading6":
		tag = "h6"
	case "Heading7":
		tag = "h6"
	default:
		tag = "div"
	}
	return tag
}

func seekText(dec *xml.Decoder) (string, error) {
	for {
		token, err := dec.Token()
		if err != nil {
			return "", err
		}
		switch tt := token.(type) {
		case xml.CharData:
			return string(tt), nil
		case xml.EndElement:
			return "", nil
		}
	}
}

func seekNextTag(dec *xml.Decoder, tag string) error {
	for {
		token, err := dec.Token()
		if err != nil {
			return err
		}
		t, ok := token.(xml.StartElement)
		if !ok {
			continue
		}
		if t.Name.Local != tag {
			continue
		}
		break
	}
	return nil
}
