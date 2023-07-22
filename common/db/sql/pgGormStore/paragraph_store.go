package pgGormStore

import (
	"context"
	"github.com/audetv/book-parser/common/app/repos/paragraph"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

var _ paragraph.ParagraphStore = &Paragraphs{}

type DBPgParagraphs []*DBPgParagraph

type DBPgParagraph struct {
	ID        uint `gorm:"primaryKey"`
	Uuid      uuid.UUID
	BookID    uint
	BookName  string
	Text      string
	Position  int
	Length    int
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type Paragraphs struct {
	db *gorm.DB
}

func NewParagraphs(dsn string) (*Paragraphs, error) {
	// Отключение авто транзакций для таблицы db_pg_paragraphs
	// иначе, при медленных запросах (вставках) более 200 мс происходил откат rollback
	// и данные не попадали в БД. Не удалось установить закономерность, при каждом новом прогоне,
	// это могли быть новые данные, разные книги, т.е. это напрямую не связанно с какими либо конкретными параграфами.
	// Включен режим логгера показывать только ошибки, предупреждения о запросах более 200 мс отключены
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(logger.Error),
	})

	db.AutoMigrate(&DBPgParagraph{})

	if err != nil {
		return nil, err
	}
	ps := &Paragraphs{
		db: db,
	}
	return ps, nil
}

func (ps *Paragraphs) Create(ctx context.Context, p *paragraph.Paragraph) error {
	dbParagraph := DBPgParagraph{
		Uuid:      p.Uuid,
		BookID:    p.BookID,
		BookName:  p.BookName,
		Text:      p.Text,
		Position:  p.Position,
		Length:    p.Length,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedAt: nil,
	}

	result := ps.db.WithContext(ctx).Create(&dbParagraph)

	return result.Error
}

func (ps *Paragraphs) BulkInsert(ctx context.Context, paragraphs []paragraph.Paragraph, batchSize int) error {
	var dbPars DBPgParagraphs
	for _, p := range paragraphs {
		dbParagraph := DBPgParagraph{
			Uuid:      p.Uuid,
			BookID:    p.BookID,
			BookName:  p.BookName,
			Text:      p.Text,
			Position:  p.Position,
			Length:    p.Length,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			DeletedAt: nil,
		}
		dbPars = append(dbPars, &dbParagraph)
	}
	result := ps.db.WithContext(ctx).CreateInBatches(dbPars, batchSize)
	return result.Error
}
