package paragraph

import (
	"context"
	"fmt"
)

// PrepareParagraphs срез подготовленных параграфов книги
type PrepareParagraphs []Paragraph

// Paragraph параграф книги
type Paragraph struct {
	BookID   uint
	BookName string
	Text     string
	Position int
	Length   int
}

type ParagraphStore interface {
	Create(ctx context.Context, paragraph *Paragraph) error
	BulkInsert(ctx context.Context, paragraphs []Paragraph, batchSize int) error
}

type Paragraphs struct {
	store ParagraphStore
}

func NewParagraphs(store ParagraphStore) *Paragraphs {
	return &Paragraphs{
		store: store,
	}
}

func (ps *Paragraphs) Create(ctx context.Context, paragraph *Paragraph) error {

	err := ps.store.Create(ctx, paragraph)

	if err != nil {
		return fmt.Errorf("create paragraph error: %w", err)
	}
	return nil
}

func (ps *Paragraphs) BulkInsert(ctx context.Context, paragraphs []Paragraph, batchSize int) error {
	err := ps.store.BulkInsert(ctx, paragraphs, batchSize)
	if err != nil {
		return fmt.Errorf("create paragraph error: %w", err)
	}
	return nil
}
