package book

import (
	"context"
	"fmt"
)

// Book книга
type Book struct {
	ID       uint
	Name     string
	Filename string
}

type BookStore interface {
	Create(ctx context.Context, b Book) (uint, error)
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
	id, err := bs.store.Create(ctx, b)
	if err != nil {
		return nil, fmt.Errorf("create book error: %w", err)
	}
	b.ID = id
	return &b, nil
}
