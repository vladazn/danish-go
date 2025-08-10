package classroom

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/vladazn/danish/app/classroom/private/mocks"
	"github.com/vladazn/danish/app/model"
	"github.com/vladazn/danish/common/userid"
)

// MockRandomWrapper wraps MockRand to satisfy the *rand.Random requirement
type MockRandomWrapper struct {
	*mocks.MockRand
}

func TestWordPool_GetBatch(t *testing.T) {
	userId := "test-user"
	now := time.Now()
	vocabId1 := uuid.New()
	vocabId2 := uuid.New()
	formId1 := uuid.New()
	formId2 := uuid.New()

	existingPool := &model.Pool{
		CreatedAt: now.Add(-1 * time.Hour), // Recent pool
		Vocabs: []model.Vocab{
			{
				Id:           vocabId1,
				Definition:   "test word 1",
				PartOfSpeech: model.PartOfSpeechNoun,
				Forms: []model.VocabForm{
					{Id: formId1, Value: "test1", Form: "singular"},
				},
			},
			{
				Id:           vocabId2,
				Definition:   "test word 2",
				PartOfSpeech: model.PartOfSpeechVerb,
				Forms: []model.VocabForm{
					{Id: formId2, Value: "test2", Form: "present"},
				},
			},
		},
	}

	expiredPool := &model.Pool{
		CreatedAt: now.Add(-6 * time.Hour), // Expired pool (>5 hours)
		Vocabs:    []model.Vocab{},
	}

	tests := []struct {
		name        string
		setupMock   func(*mocks.MockFirestore, *mocks.MockRand)
		expectedLen int
		expectError bool
	}{
		{
			name: "successfully get batch from existing pool",
			setupMock: func(mockStore *mocks.MockFirestore, mockRand *mocks.MockRand) {
				mockStore.EXPECT().
					FetchUserPool(gomock.Any(), userId).
					Return(existingPool, nil)

				// Expect shuffle to be called for batch creation
				mockRand.EXPECT().
					Shuffle(gomock.Any(), gomock.Any()).
					Do(func(n int, f func(i, j int)) {
						// Simulate shuffle by calling the swap function
						for i := 0; i < n/2; i++ {
							f(i, n-1-i)
						}
					})
			},
			expectedLen: 2,
			expectError: false,
		},
		{
			name: "build new pool when existing pool is expired",
			setupMock: func(mockStore *mocks.MockFirestore, mockRand *mocks.MockRand) {
				mockStore.EXPECT().
					FetchUserPool(gomock.Any(), userId).
					Return(expiredPool, nil)
				mockStore.EXPECT().
					FetchUserVocabulary(gomock.Any(), userId).
					Return([]model.Vocab{
						{
							Id:           vocabId1,
							Definition:   "test word 1",
							PartOfSpeech: model.PartOfSpeechNoun,
							Forms: []model.VocabForm{
								{Id: formId1, Value: "test1", Form: "singular"},
							},
						},
					}, nil)
				mockStore.EXPECT().
					UpdatePool(gomock.Any(), userId, gomock.Any()).
					Return(nil)

				// Expect shuffle to be called for batch creation
				mockRand.EXPECT().
					Shuffle(gomock.Any(), gomock.Any()).
					Do(func(n int, f func(i, j int)) {
						// Simulate shuffle by calling the swap function
						for i := 0; i < n/2; i++ {
							f(i, n-1-i)
						}
					})
			},
			expectedLen: 1,
			expectError: false,
		},
		{
			name: "build new pool when existing pool is nil",
			setupMock: func(mockStore *mocks.MockFirestore, mockRand *mocks.MockRand) {
				mockStore.EXPECT().
					FetchUserPool(gomock.Any(), userId).
					Return(nil, nil)
				mockStore.EXPECT().
					FetchUserVocabulary(gomock.Any(), userId).
					Return([]model.Vocab{
						{
							Id:           vocabId1,
							Definition:   "test word 1",
							PartOfSpeech: model.PartOfSpeechNoun,
							Forms: []model.VocabForm{
								{Id: formId1, Value: "test1", Form: "singular"},
							},
						},
					}, nil)
				mockStore.EXPECT().
					UpdatePool(gomock.Any(), userId, gomock.Any()).
					Return(nil)

				// Expect shuffle to be called for batch creation
				mockRand.EXPECT().
					Shuffle(gomock.Any(), gomock.Any()).
					Do(func(n int, f func(i, j int)) {
						// Simulate shuffle by calling the swap function
						for i := 0; i < n/2; i++ {
							f(i, n-1-i)
						}
					})
			},
			expectedLen: 1,
			expectError: false,
		},
		{
			name: "build new pool when existing pool is empty",
			setupMock: func(mockStore *mocks.MockFirestore, mockRand *mocks.MockRand) {
				mockStore.EXPECT().
					FetchUserPool(gomock.Any(), userId).
					Return(&model.Pool{Vocabs: []model.Vocab{}}, nil)
				mockStore.EXPECT().
					FetchUserVocabulary(gomock.Any(), userId).
					Return([]model.Vocab{
						{
							Id:           vocabId1,
							Definition:   "test word 1",
							PartOfSpeech: model.PartOfSpeechNoun,
							Forms: []model.VocabForm{
								{Id: formId1, Value: "test1", Form: "singular"},
							},
						},
					}, nil)
				mockStore.EXPECT().
					UpdatePool(gomock.Any(), userId, gomock.Any()).
					Return(nil)

				// Expect shuffle to be called for batch creation
				mockRand.EXPECT().
					Shuffle(gomock.Any(), gomock.Any()).
					Do(func(n int, f func(i, j int)) {
						// Simulate shuffle by calling the swap function
						for i := 0; i < n/2; i++ {
							f(i, n-1-i)
						}
					})
			},
			expectedLen: 1,
			expectError: false,
		},
		{
			name: "fetch pool error",
			setupMock: func(mockStore *mocks.MockFirestore, mockRand *mocks.MockRand) {
				mockStore.EXPECT().
					FetchUserPool(gomock.Any(), userId).
					Return(nil, errors.New("fetch error"))
			},
			expectedLen: 0,
			expectError: true,
		},
		{
			name: "fetch vocabulary error when building pool",
			setupMock: func(mockStore *mocks.MockFirestore, mockRand *mocks.MockRand) {
				mockStore.EXPECT().
					FetchUserPool(gomock.Any(), userId).
					Return(expiredPool, nil)
				mockStore.EXPECT().
					FetchUserVocabulary(gomock.Any(), userId).
					Return(nil, errors.New("vocabulary error"))
			},
			expectedLen: 0,
			expectError: true,
		},
		{
			name: "update pool error when building pool",
			setupMock: func(mockStore *mocks.MockFirestore, mockRand *mocks.MockRand) {
				mockStore.EXPECT().
					FetchUserPool(gomock.Any(), userId).
					Return(expiredPool, nil)
				mockStore.EXPECT().
					FetchUserVocabulary(gomock.Any(), userId).
					Return([]model.Vocab{
						{
							Id:           vocabId1,
							Definition:   "test word 1",
							PartOfSpeech: model.PartOfSpeechNoun,
							Forms: []model.VocabForm{
								{Id: formId1, Value: "test1", Form: "singular"},
							},
						},
					}, nil)
				mockStore.EXPECT().
					UpdatePool(gomock.Any(), userId, gomock.Any()).
					Return(errors.New("update error"))
			},
			expectedLen: 0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mocks.NewMockFirestore(ctrl)
			mockRand := mocks.NewMockRand(ctrl)
			tt.setupMock(mockStore, mockRand)

			pool := &WordPool{
				storage: mockStore,
				rand:    mockRand,
			}

			ctx := userid.ToCtx(context.Background(), userId)
			result, err := pool.GetBatch(ctx, userId)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Len(t, result.Vocabs, tt.expectedLen)
			}
		})
	}
}

func TestWordPool_RemoveFromPool(t *testing.T) {
	userId := "test-user"
	vocabId1 := uuid.New()
	vocabId2 := uuid.New()
	formId1 := uuid.New()
	formId2 := uuid.New()
	formId3 := uuid.New()

	// Helper function to create a deep copy of the pool
	createPoolCopy := func() *model.Pool {
		return &model.Pool{
			Vocabs: []model.Vocab{
				{
					Id:           vocabId1,
					Definition:   "test word 1",
					PartOfSpeech: model.PartOfSpeechNoun,
					Forms: []model.VocabForm{
						{Id: formId1, Value: "test1", Form: "singular"},
						{Id: formId2, Value: "test1", Form: "plural"},
					},
				},
				{
					Id:           vocabId2,
					Definition:   "test word 2",
					PartOfSpeech: model.PartOfSpeechVerb,
					Forms: []model.VocabForm{
						{Id: formId3, Value: "test2", Form: "present"},
					},
				},
			},
		}
	}

	tests := []struct {
		name          string
		vocabToRemove []model.Vocab
		setupMock     func(*mocks.MockFirestore)
		expectError   bool
	}{
		{
			name: "successfully remove forms from pool",
			vocabToRemove: []model.Vocab{
				{Id: vocabId1, Forms: []model.VocabForm{{Id: formId1}}},
				{Id: vocabId2, Forms: []model.VocabForm{{Id: formId3}}},
			},
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					FetchUserPool(gomock.Any(), userId).
					Return(createPoolCopy(), nil)
				mock.EXPECT().
					UpdatePool(gomock.Any(), userId, gomock.Any()).
					DoAndReturn(func(ctx context.Context, userId string, pool *model.Pool) error {
						// Verify the pool was updated correctly
						// When all forms are removed from vocabId2, the entire vocab is removed
						require.Len(t, pool.Vocabs, 1)

						// Find vocabId1 in the updated pool
						var vocab1 *model.Vocab
						for i := range pool.Vocabs {
							if pool.Vocabs[i].Id == vocabId1 {
								vocab1 = &pool.Vocabs[i]
								break
							}
						}

						// vocabId1 should have only formId2 remaining
						require.NotNil(t, vocab1)
						require.Len(t, vocab1.Forms, 1)
						require.Equal(t, formId2, vocab1.Forms[0].Id)

						return nil
					})
			},
			expectError: false,
		},
		{
			name: "remove entire vocab when all forms are removed",
			vocabToRemove: []model.Vocab{
				{Id: vocabId1, Forms: []model.VocabForm{{Id: formId1}, {Id: formId2}}},
			},
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					FetchUserPool(gomock.Any(), userId).
					Return(createPoolCopy(), nil)
				mock.EXPECT().
					UpdatePool(gomock.Any(), userId, gomock.Any()).
					DoAndReturn(func(ctx context.Context, userId string, pool *model.Pool) error {
						require.Len(t, pool.Vocabs, 1)
						require.Equal(t, vocabId2, pool.Vocabs[0].Id)
						require.Len(t, pool.Vocabs[0].Forms, 1)
						require.Equal(t, formId3, pool.Vocabs[0].Forms[0].Id)
						return nil
					})
			},
			expectError: false,
		},
		{
			name: "fetch pool error",
			vocabToRemove: []model.Vocab{
				{Id: vocabId1, Forms: []model.VocabForm{{Id: formId1}}},
			},
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					FetchUserPool(gomock.Any(), userId).
					Return(nil, errors.New("fetch error"))
			},
			expectError: true,
		},
		{
			name: "update pool error",
			vocabToRemove: []model.Vocab{
				{Id: vocabId1, Forms: []model.VocabForm{{Id: formId1}}},
			},
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					FetchUserPool(gomock.Any(), userId).
					Return(createPoolCopy(), nil)
				mock.EXPECT().
					UpdatePool(gomock.Any(), userId, gomock.Any()).
					Return(errors.New("update error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mocks.NewMockFirestore(ctrl)
			mockRand := mocks.NewMockRand(ctrl)
			tt.setupMock(mockStore)

			pool := &WordPool{
				storage: mockStore,
				rand:    mockRand,
			}

			ctx := userid.ToCtx(context.Background(), userId)
			err := pool.RemoveFromPool(ctx, userId, tt.vocabToRemove)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestWordPool_buildPool(t *testing.T) {
	userId := "test-user"
	now := time.Now()
	vocabId1 := uuid.New()
	vocabId2 := uuid.New()
	formId1 := uuid.New()
	formId2 := uuid.New()

	// Create a vocab that can be added to queue
	activeVocab := model.Vocab{
		Id:           vocabId1,
		Definition:   "active word",
		PartOfSpeech: model.PartOfSpeechNoun,
		Forms: []model.VocabForm{
			{Id: formId1, Value: "active", Form: "singular"},
		},
		PausedUntil: nil, // Can be added to queue
	}

	// Create a vocab that cannot be added to queue
	pausedVocab := model.Vocab{
		Id:           vocabId2,
		Definition:   "paused word",
		PartOfSpeech: model.PartOfSpeechVerb,
		Forms: []model.VocabForm{
			{Id: formId2, Value: "paused", Form: "present"},
		},
		PausedUntil: &now, // Cannot be added to queue
	}

	tests := []struct {
		name        string
		vocabs      []model.Vocab
		setupMock   func(*mocks.MockFirestore)
		expectError bool
	}{
		{
			name:   "successfully build pool with active vocabs",
			vocabs: []model.Vocab{activeVocab},
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					FetchUserVocabulary(gomock.Any(), userId).
					Return([]model.Vocab{activeVocab}, nil)
				mock.EXPECT().
					UpdatePool(gomock.Any(), userId, gomock.Any()).
					DoAndReturn(func(ctx context.Context, userId string, pool *model.Pool) error {
						require.Len(t, pool.Vocabs, 1)
						require.Equal(t, vocabId1, pool.Vocabs[0].Id)
						require.Len(t, pool.Vocabs[0].Forms, 1)
						require.Equal(t, formId1, pool.Vocabs[0].Forms[0].Id)
						return nil
					})
			},
			expectError: false,
		},
		{
			name:   "filter out paused vocabs",
			vocabs: []model.Vocab{activeVocab, pausedVocab},
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					FetchUserVocabulary(gomock.Any(), userId).
					Return([]model.Vocab{activeVocab, pausedVocab}, nil)
				mock.EXPECT().
					UpdatePool(gomock.Any(), userId, gomock.Any()).
					DoAndReturn(func(ctx context.Context, userId string, pool *model.Pool) error {
						require.Len(t, pool.Vocabs, 1)
						require.Equal(t, vocabId1, pool.Vocabs[0].Id)
						return nil
					})
			},
			expectError: false,
		},
		{
			name:   "fetch vocabulary error",
			vocabs: []model.Vocab{},
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					FetchUserVocabulary(gomock.Any(), userId).
					Return(nil, errors.New("fetch error"))
			},
			expectError: true,
		},
		{
			name:   "update pool error",
			vocabs: []model.Vocab{activeVocab},
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					FetchUserVocabulary(gomock.Any(), userId).
					Return([]model.Vocab{activeVocab}, nil)
				mock.EXPECT().
					UpdatePool(gomock.Any(), userId, gomock.Any()).
					Return(errors.New("update error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mocks.NewMockFirestore(ctrl)
			mockRand := mocks.NewMockRand(ctrl)
			tt.setupMock(mockStore)

			pool := &WordPool{
				storage: mockStore,
				rand:    mockRand,
			}

			ctx := userid.ToCtx(context.Background(), userId)
			result, err := pool.buildPool(ctx, userId)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.NotNil(t, result.CreatedAt)
			}
		})
	}
}

func TestWordPool_batchFromPool(t *testing.T) {
	vocabId1 := uuid.New()
	vocabId2 := uuid.New()
	formId1 := uuid.New()
	formId2 := uuid.New()
	formId3 := uuid.New()

	testPool := &model.Pool{
		Vocabs: []model.Vocab{
			{
				Id:           vocabId1,
				Definition:   "test word 1",
				PartOfSpeech: model.PartOfSpeechNoun,
				Forms: []model.VocabForm{
					{Id: formId1, Value: "test1", Form: "singular"},
					{Id: formId2, Value: "test1", Form: "plural"},
				},
			},
			{
				Id:           vocabId2,
				Definition:   "test word 2",
				PartOfSpeech: model.PartOfSpeechVerb,
				Forms: []model.VocabForm{
					{Id: formId3, Value: "test2", Form: "present"},
				},
			},
		},
	}

	tests := []struct {
		name           string
		batchSize      int
		expectedVocabs int
		expectedForms  int
	}{
		{
			name:           "batch size smaller than available forms",
			batchSize:      2,
			expectedVocabs: 2,
			expectedForms:  2,
		},
		{
			name:           "batch size larger than available forms",
			batchSize:      10,
			expectedVocabs: 2,
			expectedForms:  3,
		},
		{
			name:           "batch size of zero",
			batchSize:      0,
			expectedVocabs: 0,
			expectedForms:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRand := mocks.NewMockRand(ctrl)
			mockRand.EXPECT().
				Shuffle(gomock.Any(), gomock.Any()).
				Do(func(n int, f func(i, j int)) {
					// Simulate shuffle by calling the swap function
					for i := 0; i < n/2; i++ {
						f(i, n-1-i)
					}
				})

			pool := &WordPool{
				rand: mockRand,
			}

			result := pool.batchFromPool(testPool, tt.batchSize)

			require.Len(t, result.Vocabs, tt.expectedVocabs)

			totalForms := 0
			for _, vocab := range result.Vocabs {
				totalForms += len(vocab.Forms)
			}
			require.Equal(t, tt.expectedForms, totalForms)
		})
	}
}

func TestWordPool_ContextWithoutUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockFirestore(ctrl)
	mockRand := mocks.NewMockRand(ctrl)

	pool := &WordPool{
		storage: mockStore,
		rand:    mockRand,
	}

	ctx := context.Background() // No user ID in context

	// The wordpool methods don't use MustFromCtx, so they won't panic
	// They just take userId as a parameter and pass it to the storage layer
	// Set up mock expectations for normal operation

	// GetBatch will call FetchUserPool
	mockStore.EXPECT().
		FetchUserPool(gomock.Any(), "test-user").
		Return(nil, errors.New("storage error"))

	_, err := pool.GetBatch(ctx, "test-user")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to fetch pool")

	// RemoveFromPool will call FetchUserPool
	mockStore.EXPECT().
		FetchUserPool(gomock.Any(), "test-user").
		Return(nil, errors.New("storage error"))

	err = pool.RemoveFromPool(ctx, "test-user", []model.Vocab{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed get user pool")

	// buildPool will call FetchUserVocabulary
	mockStore.EXPECT().
		FetchUserVocabulary(gomock.Any(), "test-user").
		Return(nil, errors.New("storage error"))

	_, err = pool.buildPool(ctx, "test-user")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to fetch vocab to build pool")
}
