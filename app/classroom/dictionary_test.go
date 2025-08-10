package classroom

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/vladazn/danish/app/classroom/private/mocks"
	"github.com/vladazn/danish/app/model"
	"github.com/vladazn/danish/common/userid"
)

func TestNewDictionary(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	params := NewDictionaryParams{
		Store: &FirebaseStore{},
	}

	dict := NewDictionary(params)
	require.NotNil(t, dict)
	require.NotNil(t, dict.storage)
}

func TestDictionary_AddWord(t *testing.T) {
	tests := []struct {
		name        string
		vocab       model.Vocab
		setupMock   func(*mocks.MockFirestore)
		expectError bool
	}{
		{
			name: "successfully add word with new IDs",
			vocab: model.Vocab{
				Id:           uuid.Nil,
				Definition:   "test word",
				PartOfSpeech: model.PartOfSpeechNoun,
				Forms: []model.VocabForm{
					{Id: uuid.Nil, Value: "test", Form: "singular"},
				},
			},
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					AddVocabulary(gomock.Any(), "test-user", gomock.Any()).
					Return(nil)
			},
			expectError: false,
		},
		{
			name: "successfully add word with existing IDs",
			vocab: model.Vocab{
				Id:           uuid.New(),
				Definition:   "existing word",
				PartOfSpeech: model.PartOfSpeechVerb,
				Forms: []model.VocabForm{
					{Id: uuid.New(), Value: "existing", Form: "present"},
				},
			},
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					AddVocabulary(gomock.Any(), "test-user", gomock.Any()).
					Return(nil)
			},
			expectError: false,
		},
		{
			name: "storage error",
			vocab: model.Vocab{
				Id:           uuid.Nil,
				Definition:   "error word",
				PartOfSpeech: model.PartOfSpeechAdjective,
				Forms: []model.VocabForm{
					{Id: uuid.Nil, Value: "error", Form: "base"},
				},
			},
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					AddVocabulary(gomock.Any(), "test-user", gomock.Any()).
					Return(errors.New("test error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mocks.NewMockFirestore(ctrl)
			tt.setupMock(mockStore)

			dict := &Dictionary{storage: mockStore}
			ctx := userid.ToCtx(context.Background(), "test-user")

			result, err := dict.AddWord(ctx, tt.vocab)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)

				// Check that IDs were generated if they were nil
				if tt.vocab.Id == uuid.Nil {
					require.NotEqual(t, uuid.Nil, result.Id)
				} else {
					require.Equal(t, tt.vocab.Id, result.Id)
				}

				// Check that form IDs were generated if they were nil
				for i, form := range tt.vocab.Forms {
					if form.Id == uuid.Nil {
						require.NotEqual(t, uuid.Nil, result.Forms[i].Id)
					} else {
						require.Equal(t, form.Id, result.Forms[i].Id)
					}
				}
			}
		})
	}
}

func TestDictionary_UpdateWord(t *testing.T) {
	tests := []struct {
		name        string
		vocab       model.Vocab
		setupMock   func(*mocks.MockFirestore)
		expectError bool
	}{
		{
			name: "successfully update word",
			vocab: model.Vocab{
				Id:           uuid.New(),
				Definition:   "updated word",
				PartOfSpeech: model.PartOfSpeechNoun,
				Forms: []model.VocabForm{
					{Id: uuid.New(), Value: "updated", Form: "singular"},
				},
			},
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					AddVocabulary(gomock.Any(), "test-user", gomock.Any()).
					Return(nil)
			},
			expectError: false,
		},
		{
			name: "update word with nil IDs",
			vocab: model.Vocab{
				Id:           uuid.Nil,
				Definition:   "nil id word",
				PartOfSpeech: model.PartOfSpeechVerb,
				Forms: []model.VocabForm{
					{Id: uuid.Nil, Value: "nil", Form: "present"},
				},
			},
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					AddVocabulary(gomock.Any(), "test-user", gomock.Any()).
					Return(nil)
			},
			expectError: false,
		},
		{
			name: "storage error on update",
			vocab: model.Vocab{
				Id:           uuid.New(),
				Definition:   "error word",
				PartOfSpeech: model.PartOfSpeechAdjective,
				Forms: []model.VocabForm{
					{Id: uuid.New(), Value: "error", Form: "base"},
				},
			},
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					AddVocabulary(gomock.Any(), "test-user", gomock.Any()).
					Return(errors.New("test error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mocks.NewMockFirestore(ctrl)
			tt.setupMock(mockStore)

			dict := &Dictionary{storage: mockStore}
			ctx := userid.ToCtx(context.Background(), "test-user")

			result, err := dict.UpdateWord(ctx, tt.vocab)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)

				// Check that IDs were generated if they were nil
				if tt.vocab.Id == uuid.Nil {
					require.NotEqual(t, uuid.Nil, result.Id)
				} else {
					require.Equal(t, tt.vocab.Id, result.Id)
				}

				// Check that form IDs were generated if they were nil
				for i, form := range tt.vocab.Forms {
					if form.Id == uuid.Nil {
						require.NotEqual(t, uuid.Nil, result.Forms[i].Id)
					} else {
						require.Equal(t, form.Id, result.Forms[i].Id)
					}
				}
			}
		})
	}
}

func TestDictionary_RegisterProgress(t *testing.T) {
	userId := "test-user"
	vocabId1 := uuid.New()
	vocabId2 := uuid.New()
	formId1 := uuid.New()
	formId2 := uuid.New()

	existingVocab1 := &model.Vocab{
		Id:           vocabId1,
		Definition:   "existing word 1",
		PartOfSpeech: model.PartOfSpeechNoun,
		Forms: []model.VocabForm{
			{Id: formId1, Value: "existing1", Form: "singular"},
		},
	}

	existingVocab2 := &model.Vocab{
		Id:           vocabId2,
		Definition:   "existing word 2",
		PartOfSpeech: model.PartOfSpeechVerb,
		Forms: []model.VocabForm{
			{Id: formId2, Value: "existing2", Form: "present"},
		},
	}

	tests := []struct {
		name            string
		withoutMistakes []model.Vocab
		withMistakes    []model.Vocab
		setupMock       func(*mocks.MockFirestore)
		expectError     bool
	}{
		{
			name: "successfully register progress",
			withoutMistakes: []model.Vocab{
				{Id: vocabId1, Forms: []model.VocabForm{{Id: formId1}}},
			},
			withMistakes: []model.Vocab{
				{Id: vocabId2, Forms: []model.VocabForm{{Id: formId2}}},
			},
			setupMock: func(mock *mocks.MockFirestore) {
				// Expect GetVocab calls
				mock.EXPECT().
					GetVocab(gomock.Any(), userId, vocabId1).
					Return(existingVocab1, nil)
				mock.EXPECT().
					GetVocab(gomock.Any(), userId, vocabId2).
					Return(existingVocab2, nil)

				// Expect AddVocabulary calls for both vocabs
				mock.EXPECT().
					AddVocabulary(gomock.Any(), userId, gomock.Any()).
					Return(nil)
				mock.EXPECT().
					AddVocabulary(gomock.Any(), userId, gomock.Any()).
					Return(nil)
			},
			expectError: false,
		},
		{
			name: "vocab not found",
			withoutMistakes: []model.Vocab{
				{Id: vocabId1, Forms: []model.VocabForm{{Id: formId1}}},
			},
			withMistakes: []model.Vocab{},
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					GetVocab(gomock.Any(), userId, vocabId1).
					Return(nil, nil) // vocab not found
			},
			expectError: false, // should continue without error
		},
		{
			name: "get vocab error",
			withoutMistakes: []model.Vocab{
				{Id: vocabId1, Forms: []model.VocabForm{{Id: formId1}}},
			},
			withMistakes: []model.Vocab{},
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					GetVocab(gomock.Any(), userId, vocabId1).
					Return(nil, errors.New("test error"))
			},
			expectError: true,
		},
		{
			name: "add vocabulary error",
			withoutMistakes: []model.Vocab{
				{Id: vocabId1, Forms: []model.VocabForm{{Id: formId1}}},
			},
			withMistakes: []model.Vocab{},
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					GetVocab(gomock.Any(), userId, vocabId1).
					Return(existingVocab1, nil)
				mock.EXPECT().
					AddVocabulary(gomock.Any(), userId, gomock.Any()).
					Return(errors.New("test error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mocks.NewMockFirestore(ctrl)
			tt.setupMock(mockStore)

			dict := &Dictionary{storage: mockStore}
			ctx := userid.ToCtx(context.Background(), userId)

			err := dict.RegisterProgress(ctx, tt.withoutMistakes, tt.withMistakes)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDictionary_RemoveWord(t *testing.T) {
	vocabId := uuid.New()
	userId := "test-user"

	tests := []struct {
		name        string
		vocabId     uuid.UUID
		setupMock   func(*mocks.MockFirestore)
		expectError bool
	}{
		{
			name:    "successfully remove word",
			vocabId: vocabId,
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					RemoveVocabulary(gomock.Any(), userId, vocabId).
					Return(nil)
			},
			expectError: false,
		},
		{
			name:    "storage error on remove",
			vocabId: vocabId,
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					RemoveVocabulary(gomock.Any(), userId, vocabId).
					Return(errors.New("test error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mocks.NewMockFirestore(ctrl)
			tt.setupMock(mockStore)

			dict := &Dictionary{storage: mockStore}
			ctx := userid.ToCtx(context.Background(), userId)

			err := dict.RemoveWord(ctx, tt.vocabId)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDictionary_GetAllWords(t *testing.T) {
	userId := "test-user"
	expectedVocabs := []model.Vocab{
		{
			Id:           uuid.New(),
			Definition:   "word 1",
			PartOfSpeech: model.PartOfSpeechNoun,
			Forms: []model.VocabForm{
				{Id: uuid.New(), Value: "word1", Form: "singular"},
			},
		},
		{
			Id:           uuid.New(),
			Definition:   "word 2",
			PartOfSpeech: model.PartOfSpeechVerb,
			Forms: []model.VocabForm{
				{Id: uuid.New(), Value: "word2", Form: "present"},
			},
		},
	}

	tests := []struct {
		name        string
		setupMock   func(*mocks.MockFirestore)
		expected    []model.Vocab
		expectError bool
	}{
		{
			name: "successfully get all words",
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					FetchUserVocabulary(gomock.Any(), userId).
					Return(expectedVocabs, nil)
			},
			expected:    expectedVocabs,
			expectError: false,
		},
		{
			name: "storage error",
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					FetchUserVocabulary(gomock.Any(), userId).
					Return(nil, errors.New("test error"))
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "empty vocabulary",
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					FetchUserVocabulary(gomock.Any(), userId).
					Return([]model.Vocab{}, nil)
			},
			expected:    []model.Vocab{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mocks.NewMockFirestore(ctrl)
			tt.setupMock(mockStore)

			dict := &Dictionary{storage: mockStore}
			ctx := userid.ToCtx(context.Background(), userId)

			result, err := dict.GetAllWords(ctx)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestDictionary_RegisterProgress_ConcurrentUpdates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockFirestore(ctrl)
	userId := "test-user"
	vocabId := uuid.New()
	formId := uuid.New()

	existingVocab := &model.Vocab{
		Id:           vocabId,
		Definition:   "test word",
		PartOfSpeech: model.PartOfSpeechNoun,
		Forms: []model.VocabForm{
			{Id: formId, Value: "test", Form: "singular"},
		},
	}

	// Setup mock expectations for concurrent calls
	mockStore.EXPECT().
		GetVocab(gomock.Any(), userId, vocabId).
		Return(existingVocab, nil).
		Times(1) // Called once since it's the same vocab ID

	mockStore.EXPECT().
		AddVocabulary(gomock.Any(), userId, gomock.Any()).
		Return(nil).
		Times(1) // Only called once since it's the same vocab

	dict := &Dictionary{storage: mockStore}
	ctx := userid.ToCtx(context.Background(), userId)

	// Test concurrent updates to the same vocab
	withoutMistakes := []model.Vocab{
		{Id: vocabId, Forms: []model.VocabForm{{Id: formId}}},
	}
	withMistakes := []model.Vocab{
		{Id: vocabId, Forms: []model.VocabForm{{Id: formId}}},
	}

	err := dict.RegisterProgress(ctx, withoutMistakes, withMistakes)
	require.NoError(t, err)
}

func TestDictionary_ContextWithoutUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockFirestore(ctrl)
	dict := &Dictionary{storage: mockStore}
	ctx := context.Background() // No user ID in context

	// These should panic due to missing user ID
	require.Panics(t, func() {
		dict.AddWord(ctx, model.Vocab{})
	})

	require.Panics(t, func() {
		dict.UpdateWord(ctx, model.Vocab{})
	})

	require.Panics(t, func() {
		dict.RegisterProgress(ctx, []model.Vocab{}, []model.Vocab{})
	})

	require.Panics(t, func() {
		dict.RemoveWord(ctx, uuid.New())
	})

	require.Panics(t, func() {
		dict.GetAllWords(ctx)
	})
}
