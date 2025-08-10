package classroom

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/fx"

	"github.com/vladazn/danish/app/model"
	"github.com/vladazn/danish/common/rand"
)

//go:generate go run go.uber.org/mock/mockgen@latest -source=wordpool.go -destination=private/mocks/rand_mock.go -package=mocks
type Rand interface {
	IntN(n int) int
	Shuffle(n int, f func(i, j int))
}

type WordPool struct {
	storage Firestore
	rand    Rand
}

type NewWordPoolParams struct {
	fx.In
	Rand  *rand.Random
	Store *FirebaseStore
}

func NewWordPool(p NewWordPoolParams) *WordPool {
	return &WordPool{
		rand:    p.Rand,
		storage: p.Store,
	}
}

func (wp *WordPool) buildPool(ctx context.Context, userId string) (*model.Pool, error) {
	pool := &model.Pool{}
	vocabs, err := wp.storage.FetchUserVocabulary(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch vocab to build pool: %w", err)
	}
	now := time.Now()

	filtered := []model.Vocab{}

	for _, v := range vocabs {
		if !v.CanBeAddedToQueue(now) {
			continue
		}
		var filteredForms []model.VocabForm
		for _, form := range v.Forms {
			if form.CanBeAddedToQueue(now) {
				filteredForms = append(filteredForms, form)
			}
		}

		if len(filteredForms) > 0 {
			v.Forms = filteredForms
			filtered = append(filtered, v)
		}
	}

	pool.Vocabs = filtered
	pool.CreatedAt = time.Now()

	err = wp.storage.UpdatePool(ctx, userId, pool)
	if err != nil {
		return nil, fmt.Errorf("failed to update pool: %w", err)
	}

	return pool, nil
}

func (wp *WordPool) GetBatch(ctx context.Context, userId string) (*model.Batch, error) {
	pool, err := wp.storage.FetchUserPool(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pool: %w", err)
	}

	if pool == nil || len(pool.Vocabs) == 0 || pool.CreatedAt.Before(time.Now().Add(-5*time.Hour)) {
		pool, err = wp.buildPool(ctx, userId)
		if err != nil {
			return nil, fmt.Errorf("failed to build pool: %w", err)
		}
	}

	batch := wp.batchFromPool(pool, 10)

	return &batch, nil
}

func (wp *WordPool) RemoveFromPool(ctx context.Context, userId string, vocab []model.Vocab) error {
	pool, err := wp.storage.FetchUserPool(ctx, userId)
	if err != nil {
		return fmt.Errorf("failed get user pool: %w", err)
	}
	wp.removeVocabFromPool(pool, vocab)

	err = wp.storage.UpdatePool(ctx, userId, pool)
	if err != nil {
		return fmt.Errorf("failed to save updated pool: %w", err)
	}

	return nil
}

func (wp *WordPool) removeVocabFromPool(pool *model.Pool, vocabToRemove []model.Vocab) {
	formsToRemove := make(map[uuid.UUID]map[uuid.UUID]bool)
	for _, vocab := range vocabToRemove {
		if _, exists := formsToRemove[vocab.Id]; !exists {
			formsToRemove[vocab.Id] = make(map[uuid.UUID]bool)
		}
		for _, form := range vocab.Forms {
			formsToRemove[vocab.Id][form.Id] = true
		}
	}

	var updatedVocabs []model.Vocab

	for _, vocab := range pool.Vocabs {
		if formMap, exists := formsToRemove[vocab.Id]; exists {
			var filteredForms []model.VocabForm
			for _, form := range vocab.Forms {
				if !formMap[form.Id] {
					filteredForms = append(filteredForms, form)
				}
			}
			if len(filteredForms) > 0 {
				vocab.Forms = filteredForms
				updatedVocabs = append(updatedVocabs, vocab)
			}
		} else {
			updatedVocabs = append(updatedVocabs, vocab)
		}
	}

	pool.Vocabs = updatedVocabs
}

func (wp *WordPool) batchFromPool(pool *model.Pool, n int) model.Batch {
	type formWithParent struct {
		ParentVocab model.Vocab
		Form        model.VocabForm
	}

	var allForms []formWithParent

	for _, vocab := range pool.Vocabs {
		for _, form := range vocab.Forms {
			allForms = append(allForms, formWithParent{
				ParentVocab: vocab,
				Form:        form,
			})
		}
	}

	wp.rand.Shuffle(len(allForms), func(i, j int) {
		allForms[i], allForms[j] = allForms[j], allForms[i]
	})

	if len(allForms) < n {
		n = len(allForms)
	}
	selected := allForms[:n]

	vocabMap := make(map[uuid.UUID]model.Vocab)
	for _, item := range selected {
		v, exists := vocabMap[item.ParentVocab.Id]
		if !exists {
			v = model.Vocab{
				Id:           item.ParentVocab.Id,
				Definition:   item.ParentVocab.Definition,
				PartOfSpeech: item.ParentVocab.PartOfSpeech,
				Forms:        []model.VocabForm{},
			}
		}
		v.Forms = append(v.Forms, item.Form)
		vocabMap[item.ParentVocab.Id] = v
	}

	var batch model.Batch
	for _, vocab := range vocabMap {
		batch.Vocabs = append(batch.Vocabs, vocab)
	}

	return batch
}
