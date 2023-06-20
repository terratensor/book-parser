package book

import (
	"context"
	"fmt"
	"github.com/google/uuid"
)

// Book книга
type Book struct {
	ID       uuid.UUID
	Name     string
	Filename string
}

type BookStore interface {
	Create(ctx context.Context, b Book) (*uuid.UUID, error)
}

type Books struct {
	store BookStore
}

func NewBooks(store BookStore) *Books {
	return &Books{
		store: store,
	}
}

func (bs *Books) Create(ctx context.Context, b Book) (*Book, error) {
	b.ID = uuid.New()
	id, err := bs.store.Create(ctx, b)
	if err != nil {
		return nil, fmt.Errorf("create book error: %w", err)
	}
	b.ID = *id
	return &b, nil
}
