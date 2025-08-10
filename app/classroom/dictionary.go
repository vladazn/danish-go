package classroom

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/fx"
	"golang.org/x/sync/errgroup"

	"github.com/vladazn/danish/app/model"
	"github.com/vladazn/danish/common/userid"
)

type NewDictionaryParams struct {
	fx.In
	Store *FirebaseStore
}

func NewDictionary(p NewDictionaryParams) *Dictionary {
	return &Dictionary{
		storage: p.Store,
	}
}

type Dictionary struct {
	storage Firestore
}

func (d *Dictionary) AddWord(ctx context.Context, vocab model.Vocab) (model.Vocab, error) {
	if vocab.Id == uuid.Nil {
		vocab.Id = uuid.New()
	}

	for i, form := range vocab.Forms {
		if form.Id == uuid.Nil {
			vocab.Forms[i].Id = uuid.New()
		}
	}
	return vocab, d.storage.AddVocabulary(ctx, userid.MustFromCtx(ctx), vocab)
}

func (d *Dictionary) UpdateWord(ctx context.Context, vocab model.Vocab) (model.Vocab, error) {
	if vocab.Id == uuid.Nil {
		vocab.Id = uuid.New()
	}
	for i, form := range vocab.Forms {
		if form.Id == uuid.Nil {
			vocab.Forms[i].Id = uuid.New()
		}
	}
	return vocab, d.storage.AddVocabulary(ctx, userid.MustFromCtx(ctx), vocab)
}

func (d *Dictionary) RegisterProgress(
	ctx context.Context, withoutMistakes []model.Vocab, withMistakes []model.Vocab,
) error {
	userId := userid.MustFromCtx(ctx)
	m := map[uuid.UUID]*model.Vocab{}

	for _, vocabToUpdate := range withoutMistakes {
		if _, ok := m[vocabToUpdate.Id]; !ok {
			vocab, err := d.storage.GetVocab(ctx, userId, vocabToUpdate.Id)
			if err != nil {
				return fmt.Errorf("could not find vocab: %w", err)
			}
			if vocab == nil {
				continue
			}
			m[vocabToUpdate.Id] = vocab
		}
		for _, form := range vocabToUpdate.Forms {
			m[vocabToUpdate.Id].UpdateFormSuccess(form, time.Now())
		}
	}

	for _, vocabToUpdate := range withMistakes {
		if _, ok := m[vocabToUpdate.Id]; !ok {
			vocab, err := d.storage.GetVocab(ctx, userId, vocabToUpdate.Id)
			if err != nil {
				return fmt.Errorf("could not find vocab: %w", err)
			}
			if vocab == nil {
				continue
			}
			m[vocabToUpdate.Id] = vocab
		}

		for _, form := range vocabToUpdate.Forms {
			m[vocabToUpdate.Id].UpdateFormSuccess(form, time.Now())
		}
	}

	eg, egCtx := errgroup.WithContext(ctx)

	for _, updatedVocabs := range m {
		eg.Go(func() error {
			return d.storage.AddVocabulary(egCtx, userId, *updatedVocabs)
		})
	}

	err := eg.Wait()
	if err != nil {
		return fmt.Errorf("could not update vocab progress: %w", err)
	}

	return nil
}

func (d *Dictionary) RemoveWord(ctx context.Context, vocabId uuid.UUID) error {
	return d.storage.RemoveVocabulary(ctx, userid.MustFromCtx(ctx), vocabId)
}

func (d *Dictionary) GetAllWords(ctx context.Context) ([]model.Vocab, error) {
	return d.storage.FetchUserVocabulary(ctx, userid.MustFromCtx(ctx))
}
