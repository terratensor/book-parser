package pgstore

import (
	"context"
	"database/sql"
	"github.com/audetv/book-parser/develop/app/repos/book"
	"github.com/audetv/book-parser/develop/app/repos/paragraph"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v4/stdlib" // Postgresql driver
	"strconv"
	"strings"
	"time"
)

var _ book.BookStore = &Books{}

type DBPgParagraph struct {
	UUID      uuid.UUID  `db:"uuid"`
	BookUUID  uuid.UUID  `db:"book_uuid"`
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
		UUID:      p.ID,
		BookUUID:  p.BookID,
		Text:      p.Text,
		Position:  p.Position,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err := ps.db.ExecContext(ctx, `INSERT INTO book_paragraphs
    (uuid, book_uuid, text, position, created_at, updated_at, deleted_at)
    values ($1, $2, $3, $4, $5, $6, $7)`,
		dbp.UUID,
		dbp.BookUUID,
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

func (ps *Paragraphs) BulkInsert(ctx context.Context, paragraphs *paragraph.PrepareParagraphs) error {
	var values []interface{}
	for _, row := range *paragraphs {
		values = append(values, row.ID, row.BookID, row.Text, row.Position, time.Now(), time.Now(), nil)
	}

	stmt := Build(`INSERT INTO test(uuid, book_uuid, text, position, created_at, updated_at, deleted_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, len(*paragraphs))

	_, err := ps.db.ExecContext(ctx, stmt)
	if err != nil {
		return err
	}

	return nil
}

func Build(stmt string, len int) string {
	beforVals := stmt[:strings.IndexByte(stmt, '?')-1]
	afterVals := stmt[strings.LastIndexByte(stmt, '?')+2:]
	vals := stmt[strings.IndexByte(stmt, '?')-1 : strings.LastIndexByte(stmt, '?')+2]
	vals += strings.Repeat(","+vals, len)
	stmt = beforVals + vals + afterVals
	n := 0
	for strings.IndexByte(stmt, '?') != -1 {
		n++
		param := "$" + strconv.Itoa(n)
		stmt = strings.Replace(stmt, "?", param, 1)
	}
	return stmt
}
