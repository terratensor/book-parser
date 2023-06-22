package pgGormStore

import (
	"context"
	"github.com/audetv/book-parser/common/app/repos/book"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

var _ book.BookStore = &Books{}

type DBPgBook struct {
	ID        uint `gorm:"primaryKey"`
	Name      string
	Filename  string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type Books struct {
	db *gorm.DB
}

func NewBooks(dsn string) (*Books, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	db.AutoMigrate(&DBPgBook{})

	if err != nil {
		return nil, err
	}
	bs := &Books{
		db: db,
	}
	return bs, nil
}

func (bs *Books) Create(ctx context.Context, b book.Book) (uint, error) {
	dbBook := DBPgBook{
		Name:      b.Name,
		Filename:  b.Filename,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedAt: nil,
	}

	result := bs.db.WithContext(ctx).Create(&dbBook)

	return dbBook.ID, result.Error
}
