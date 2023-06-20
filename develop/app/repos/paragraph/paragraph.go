package paragraph

import (
	"context"
	"fmt"
	"github.com/google/uuid"
)

// PrepareParagraphs срез подготовленных параграфов книги
type PrepareParagraphs []Paragraph

// Paragraph параграф книги
type Paragraph struct {
	ID       uuid.UUID
	BookID   uuid.UUID
	Text     string
	Position int
}

type ParagraphStore interface {
	Create(ctx context.Context, paragraph *Paragraph) error
	BulkInsert(ctx context.Context, paragraphs *PrepareParagraphs) error
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
