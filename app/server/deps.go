package server

import (
	"context"

	"github.com/google/uuid"

	"github.com/vladazn/danish/app/model"
)

type Dictionary interface {
	AddWord(ctx context.Context, vocab model.Vocab) error
	UpdateWord(ctx context.Context, vocab model.Vocab) error
	RemoveWord(ctx context.Context, vocabId uuid.UUID) error
	GetAllWords(ctx context.Context) ([]model.Vocab, error)
}
