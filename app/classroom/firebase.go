package classroom

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/fx"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/vladazn/danish/app/model"
	"github.com/vladazn/danish/app/storage"
)

//go:generate go run go.uber.org/mock/mockgen@latest -source=firebase.go -destination=private/mocks/firebase_mock.go -package=mocks
type Firestore interface {
	FetchUserVocabulary(ctx context.Context, userId string) ([]model.Vocab, error)
	AddVocabulary(ctx context.Context, userId string, vocab model.Vocab) error
	RemoveVocabulary(ctx context.Context, userId string, vocabId uuid.UUID) error
	UpdatePool(ctx context.Context, userId string, pool *model.Pool) error
	FetchUserPool(ctx context.Context, userId string) (*model.Pool, error)
	GetVocab(ctx context.Context, userId string, vocabId uuid.UUID) (*model.Vocab, error)
	GetMultipleVocabs(ctx context.Context, userId string, vocabIds []uuid.UUID) ([]model.Vocab, error)
	SetVocabSet(ctx context.Context, userId string, vocabSet model.VocabSet) error
	GetVocabSet(ctx context.Context, userId string, vocabSetId uuid.UUID) (*model.VocabSet, error)
	FetchUserVocabSets(ctx context.Context, userId string) ([]model.VocabSet, error)
	RemoveVocabSet(ctx context.Context, userId string, vocabSetId uuid.UUID) error
}

type FirebaseStore struct {
	client *storage.FirestoreClient
}

type NewFirebaseStoreParams struct {
	fx.In
	Client *storage.FirestoreClient
}

func NewFirebaseStore(p NewFirebaseStoreParams) *FirebaseStore {
	return &FirebaseStore{
		client: p.Client,
	}
}

// StorageVocabSet is a Firestore-compatible version of VocabSet
type StorageVocabSet struct {
	Id       string   `firestore:"id"`
	Name     string   `firestore:"name"`
	VocabIds []string `firestore:"vocab_ids"`
}

// toStorageVocabSet converts a VocabSet to StorageVocabSet
func toStorageVocabSet(vs model.VocabSet) StorageVocabSet {
	vocabIds := make([]string, len(vs.VocabIds))
	for i, id := range vs.VocabIds {
		vocabIds[i] = id.String()
	}

	return StorageVocabSet{
		Id:       vs.Id.String(),
		Name:     vs.Name,
		VocabIds: vocabIds,
	}
}

// fromStorageVocabSet converts a StorageVocabSet to VocabSet
func fromStorageVocabSet(svs StorageVocabSet) (*model.VocabSet, error) {
	vocabIds := make([]uuid.UUID, len(svs.VocabIds))
	for i, idStr := range svs.VocabIds {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, fmt.Errorf("invalid vocab ID %s: %w", idStr, err)
		}
		vocabIds[i] = id
	}

	id, err := uuid.Parse(svs.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid set ID %s: %w", svs.Id, err)
	}

	return &model.VocabSet{
		Id:       id,
		Name:     svs.Name,
		VocabIds: vocabIds,
	}, nil
}

func (fs *FirebaseStore) GetVocab(ctx context.Context, userId string, vocabId uuid.UUID) (*model.Vocab, error) {
	var vocab model.Vocab
	doc, err := fs.client.Client.Collection("users").Doc(userId).Collection("vocab").
		Doc(vocabId.String()).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}

	err = doc.DataTo(&vocab)
	if err != nil {
		return nil, err
	}

	return &vocab, nil
}

func (fs *FirebaseStore) FetchUserVocabulary(ctx context.Context, userId string) ([]model.Vocab, error) {
	iter := fs.client.Client.Collection("users").Doc(userId).Collection("vocab").Documents(ctx)
	defer iter.Stop()

	var results []model.Vocab
	for {
		doc, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return nil, fmt.Errorf("failed to fetch vocab iter: %w", err)
		}

		var v model.Vocab
		if err := doc.DataTo(&v); err != nil {
			continue
		}
		results = append(results, v)
	}

	return results, nil
}

func (fs *FirebaseStore) AddVocabulary(ctx context.Context, userId string, vocab model.Vocab) error {
	_, err := fs.client.Client.Collection("users").Doc(userId).Collection("vocab").
		Doc(vocab.Id.String()).Set(ctx, vocab)

	if err != nil {
		return fmt.Errorf("failed to add vocab: %w", err)
	}

	return nil
}

func (fs *FirebaseStore) RemoveVocabulary(ctx context.Context, userId string, vocabId uuid.UUID) error {
	_, err := fs.client.Client.Collection("users").Doc(userId).Collection("vocab").
		Doc(vocabId.String()).Delete(ctx)

	if err != nil {
		return fmt.Errorf("failed to remove vocab: %w", err)
	}

	return nil
}

func (fs *FirebaseStore) UpdatePool(ctx context.Context, userId string, pool *model.Pool) error {
	_, err := fs.client.Client.Collection("users").Doc(userId).Collection("pool").
		Doc("main").Set(ctx, pool)

	if err != nil {
		return fmt.Errorf("failed to update pool: %w", err)
	}

	return nil
}

func (fs *FirebaseStore) FetchUserPool(ctx context.Context, userId string) (*model.Pool, error) {
	doc, err := fs.client.Client.Collection("users").Doc(userId).Collection("pool").
		Doc("main").Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}

	var pool model.Pool
	err = doc.DataTo(&pool)
	if err != nil {
		return nil, err
	}

	return &pool, nil
}

func (fs *FirebaseStore) GetMultipleVocabs(ctx context.Context, userId string, vocabIds []uuid.UUID) ([]model.Vocab, error) {
	if len(vocabIds) == 0 {
		return []model.Vocab{}, nil
	}

	var results []model.Vocab
	collection := fs.client.Client.Collection("users").Doc(userId).Collection("vocab")

	for _, vocabId := range vocabIds {
		doc, err := collection.Doc(vocabId.String()).Get(ctx)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				continue
			}
			return nil, fmt.Errorf("failed to fetch vocab %s: %w", vocabId.String(), err)
		}

		var vocab model.Vocab
		if err := doc.DataTo(&vocab); err != nil {
			continue
		}
		results = append(results, vocab)
	}

	return results, nil
}

func (fs *FirebaseStore) SetVocabSet(ctx context.Context, userId string, vocabSet model.VocabSet) error {
	storageVocabSet := toStorageVocabSet(vocabSet)
	_, err := fs.client.Client.Collection("users").Doc(userId).Collection("sets").
		Doc(vocabSet.Id.String()).Set(ctx, storageVocabSet)

	if err != nil {
		return fmt.Errorf("failed to set vocab set: %w", err)
	}

	return nil
}

func (fs *FirebaseStore) GetVocabSet(ctx context.Context, userId string, vocabSetId uuid.UUID) (*model.VocabSet, error) {
	var storageVocabSet StorageVocabSet
	doc, err := fs.client.Client.Collection("users").Doc(userId).Collection("sets").
		Doc(vocabSetId.String()).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}

	err = doc.DataTo(&storageVocabSet)
	if err != nil {
		return nil, err
	}

	return fromStorageVocabSet(storageVocabSet)
}

func (fs *FirebaseStore) FetchUserVocabSets(ctx context.Context, userId string) ([]model.VocabSet, error) {
	iter := fs.client.Client.Collection("users").Doc(userId).Collection("sets").Documents(ctx)
	defer iter.Stop()

	results := []model.VocabSet{}
	for {
		doc, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return nil, fmt.Errorf("failed to fetch vocab sets iter: %w", err)
		}

		var storageVocabSet StorageVocabSet
		if err := doc.DataTo(&storageVocabSet); err != nil {
			continue
		}

		vocabSet, err := fromStorageVocabSet(storageVocabSet)
		if err != nil {
			continue // Skip invalid sets
		}
		results = append(results, *vocabSet)
	}

	return results, nil
}

func (fs *FirebaseStore) RemoveVocabSet(ctx context.Context, userId string, vocabSetId uuid.UUID) error {
	_, err := fs.client.Client.Collection("users").Doc(userId).Collection("sets").
		Doc(vocabSetId.String()).Delete(ctx)

	if err != nil {
		return fmt.Errorf("failed to remove vocab set: %w", err)
	}

	return nil
}
