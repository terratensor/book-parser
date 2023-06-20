package pgstore

import (
	"context"
	"database/sql"
	"github.com/audetv/book-parser/develop/app/repos/book"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v4/stdlib" // Postgresql driver
	"time"
)

var _ book.BookStore = &Books{}

type DBPgBook struct {
	UUID      uuid.UUID  `db:"uuid"`
	Name      string     `db:"name"`
	Filename  string     `db:"filename"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

type Books struct {
	db *sql.DB
}

func NewBooks(dsn string) (*Books, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}
	bs := &Books{
		db: db,
	}
	return bs, nil
}

func (bs *Books) Close() {
	bs.db.Close()
}

func (bs *Books) Create(ctx context.Context, b book.Book) (*uuid.UUID, error) {
	dbp := &DBPgBook{
		UUID:      b.ID,
		Name:      b.Name,
		Filename:  b.Filename,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err := bs.db.ExecContext(ctx, `INSERT INTO books
    (uuid, name, filename, created_at, updated_at, deleted_at)
    values ($1, $2, $3, $4, $5, $6)`,
		dbp.UUID,
		dbp.Name,
		dbp.Filename,
		dbp.CreatedAt,
		dbp.UpdatedAt,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &b.ID, nil
}