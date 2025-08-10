package classroom

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/fx"

	"github.com/vladazn/danish/app/model"
)

type SetService struct {
	storage Firestore
}

type NewSetServiceParams struct {
	fx.In
	Store *FirebaseStore
}

func NewSetService(p NewSetServiceParams) *SetService {
	return &SetService{
		storage: p.Store,
	}
}

// GetSetList retrieves all vocab sets for a user
func (ss *SetService) GetSetList(ctx context.Context, userId string) ([]model.VocabSet, error) {
	sets, err := ss.storage.FetchUserVocabSets(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user vocab sets: %w", err)
	}
	return sets, nil
}

// GetSetVocabBatch retrieves a batch of vocabulary items from a specific set
func (ss *SetService) GetSetVocabBatch(ctx context.Context, userId string, setId uuid.UUID, limit int) (*model.Batch, error) {
	// Get the vocab set
	vocabSet, err := ss.storage.GetVocabSet(ctx, userId, setId)
	if err != nil {
		return nil, fmt.Errorf("failed to get vocab set: %w", err)
	}
	if vocabSet == nil {
		return nil, fmt.Errorf("vocab set not found")
	}

	// Get the vocabulary items for the set
	vocabs, err := ss.storage.GetMultipleVocabs(ctx, userId, vocabSet.VocabIds)
	if err != nil {
		return nil, fmt.Errorf("failed to get vocab items for set: %w", err)
	}

	// Limit the number of items if requested
	if limit > 0 && len(vocabs) > limit {
		vocabs = vocabs[:limit]
	}

	return &model.Batch{
		Vocabs: vocabs,
	}, nil
}

// AddSet creates a new vocab set for a user
func (ss *SetService) AddSet(ctx context.Context, userId string, name string, vocabIds []uuid.UUID) (*model.VocabSet, error) {
	vocabSet := model.VocabSet{
		Id:       uuid.New(),
		Name:     name,
		VocabIds: vocabIds,
	}

	err := ss.storage.SetVocabSet(ctx, userId, vocabSet)
	if err != nil {
		return nil, fmt.Errorf("failed to add vocab set: %w", err)
	}

	return &vocabSet, nil
}

// UpdateSet updates an existing vocab set
func (ss *SetService) UpdateSet(ctx context.Context, userId string, setId uuid.UUID, name string, vocabIds []uuid.UUID) error {
	// Check if the set exists
	existingSet, err := ss.storage.GetVocabSet(ctx, userId, setId)
	if err != nil {
		return fmt.Errorf("failed to get existing vocab set: %w", err)
	}
	if existingSet == nil {
		return fmt.Errorf("vocab set not found")
	}

	// Update the set
	updatedSet := model.VocabSet{
		Id:       setId,
		Name:     name,
		VocabIds: vocabIds,
	}

	err = ss.storage.SetVocabSet(ctx, userId, updatedSet)
	if err != nil {
		return fmt.Errorf("failed to update vocab set: %w", err)
	}

	return nil
}

// RemoveSet removes a vocab set for a user
func (ss *SetService) RemoveSet(ctx context.Context, userId string, setId uuid.UUID) error {
	// Check if the set exists
	existingSet, err := ss.storage.GetVocabSet(ctx, userId, setId)
	if err != nil {
		return fmt.Errorf("failed to get existing vocab set: %w", err)
	}
	if existingSet == nil {
		return fmt.Errorf("vocab set not found")
	}

	// Remove the set
	err = ss.storage.RemoveVocabSet(ctx, userId, setId)
	if err != nil {
		return fmt.Errorf("failed to remove vocab set: %w", err)
	}

	return nil
}
