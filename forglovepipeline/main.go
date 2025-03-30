package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/text/encoding/charmap"
)

func main() {
	// Путь к директории с книгами
	booksDir := "./books/VPSSSR/TXT"

	// Путь для сохранения обработанных файлов
	outputDir := "./processed_books"

	// Создаем директорию для обработанных файлов, если ее нет
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Не удалось создать директорию %s: %v", outputDir, err)
	}

	// Получаем список файлов в директории
	files, err := ioutil.ReadDir(booksDir)
	if err != nil {
		log.Fatalf("Не удалось прочитать директорию %s: %v", booksDir, err)
	}

	// Создаем объединенный файл для GloVe
	combinedFile, err := os.Create(filepath.Join(outputDir, "combined.txt"))
	if err != nil {
		log.Fatalf("Не удалось создать объединенный файл: %v", err)
	}
	defer combinedFile.Close()

	// Обрабатываем каждый файл
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Пропускаем не txt файлы
		if filepath.Ext(file.Name()) != ".txt" {
			continue
		}

		// Полный путь к файлу
		filePath := filepath.Join(booksDir, file.Name())

		// Обрабатываем файл
		processedText, err := processBookFile(filePath)
		if err != nil {
			log.Printf("Ошибка при обработке файла %s: %v", filePath, err)
			continue
		}

		// Сохраняем обработанный файл
		outputPath := filepath.Join(outputDir, "processed_"+file.Name())
		if err := ioutil.WriteFile(outputPath, []byte(processedText), 0644); err != nil {
			log.Printf("Не удалось сохранить обработанный файл %s: %v", outputPath, err)
		}

		// Записываем в объединенный файл
		if _, err := combinedFile.WriteString(processedText + "\n"); err != nil {
			log.Printf("Не удалось записать в объединенный файл: %v", err)
		}

		fmt.Printf("Обработан файл: %s\n", file.Name())
	}

	fmt.Println("Обработка завершена. Результаты сохранены в", outputDir)
}

// processBookFile обрабатывает один файл книги
func processBookFile(filePath string) (string, error) {
	// Читаем файл
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("не удалось прочитать файл: %v", err)
	}

	// Конвертируем из windows-1251 в UTF-8
	decoder := charmap.Windows1251.NewDecoder()
	utf8Content, err := decoder.Bytes(content)
	if err != nil {
		return "", fmt.Errorf("ошибка конвертации в UTF-8: %v", err)
	}

	// Преобразуем в строку и обрабатываем
	text := string(utf8Content)
	text = preprocessText(text)

	return text, nil
}

// preprocessText выполняет предварительную обработку текста
func preprocessText(text string) string {
	// Приводим к нижнему регистру
	text = strings.ToLower(text)

	// Удаляем специальные символы, оставляем только буквы, цифры и пробелы
	reg := regexp.MustCompile(`[^а-яё0-9\s]`)
	text = reg.ReplaceAllString(text, " ")

	// Заменяем множественные пробелы на одинарные
	reg = regexp.MustCompile(`\s+`)
	text = reg.ReplaceAllString(text, " ")

	// Удаляем начальные и конечные пробелы
	text = strings.TrimSpace(text)

	// Разбиваем текст на токены (слова) и собираем обратно,
	// чтобы убедиться в правильном формате для GloVe
	scanner := bufio.NewScanner(bytes.NewReader([]byte(text)))
	scanner.Split(bufio.ScanWords)

	var words []string
	for scanner.Scan() {
		word := scanner.Text()
		// Можно добавить дополнительную обработку слов здесь
		if len(word) > 1 { // Игнорируем однобуквенные слова
			words = append(words, word)
		}
	}

	return strings.Join(words, " ")
}
