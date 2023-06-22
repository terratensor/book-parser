package pgstore

import (
	"context"
	"database/sql"
	"github.com/audetv/book-parser/common/app/repos/paragraph"
	_ "github.com/jackc/pgx/v4/stdlib" // Postgresql driver
	"time"
)

//var _ paragraph.ParagraphStore = &Paragraphs{}

type DBPgParagraph struct {
	ID        uint       `db:"id"`
	BookID    uint       `db:"book_id"`
	Text      string     `db:"text"`
	Position  int        `db:"position"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

type Paragraphs struct {
	db *sql.DB
}

func NewParagraphs(dsn string) (*Paragraphs, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}
	ps := &Paragraphs{
		db: db,
	}
	return ps, nil
}

func (ps *Paragraphs) Close() {
	ps.db.Close()
}

func (ps *Paragraphs) Create(ctx context.Context, p *paragraph.Paragraph) error {
	dbp := &DBPgParagraph{
		BookID:    p.BookID,
		Text:      p.Text,
		Position:  p.Position,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err := ps.db.ExecContext(ctx, `INSERT INTO book_paragraphs
    (id, book_id, text, position, created_at, updated_at, deleted_at)
    values ($1, $2, $3, $4, $5, $6, $7)`,
		dbp.ID,
		dbp.BookID,
		dbp.Text,
		dbp.Position,
		dbp.CreatedAt,
		dbp.UpdatedAt,
		nil,
	)
	if err != nil {
		return err
	}

	return nil
}

func (ps *Paragraphs) BulkInsert(ctx context.Context, paragraphs []paragraph.Paragraph, batchSize int) error {
	return nil
}
