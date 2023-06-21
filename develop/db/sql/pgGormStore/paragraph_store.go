package pgGormStore

import (
	"context"
	"github.com/audetv/book-parser/develop/app/repos/paragraph"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

var _ paragraph.ParagraphStore = &Paragraphs{}

type DBPgParagraphs []*DBPgParagraph

type DBPgParagraph struct {
	ID        uint `gorm:"primaryKey"`
	BookID    uint
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
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

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
		BookID:    p.BookID,
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
			BookID:    p.BookID,
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
