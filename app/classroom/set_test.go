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

func TestSetService_GetSetList(t *testing.T) {
	userId := "test-user"
	setId1 := uuid.New()
	setId2 := uuid.New()
	vocabId1 := uuid.New()
	vocabId2 := uuid.New()

	expectedSets := []model.VocabSet{
		{
			Id:       setId1,
			Name:     "Test Set 1",
			VocabIds: []uuid.UUID{vocabId1},
		},
		{
			Id:       setId2,
			Name:     "Test Set 2",
			VocabIds: []uuid.UUID{vocabId2},
		},
	}

	tests := []struct {
		name        string
		setupMock   func(*mocks.MockFirestore)
		expected    []model.VocabSet
		expectError bool
	}{
		{
			name: "successfully get set list",
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					FetchUserVocabSets(gomock.Any(), userId).
					Return(expectedSets, nil)
			},
			expected:    expectedSets,
			expectError: false,
		},
		{
			name: "fetch sets error",
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					FetchUserVocabSets(gomock.Any(), userId).
					Return(nil, errors.New("fetch error"))
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "empty set list",
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					FetchUserVocabSets(gomock.Any(), userId).
					Return([]model.VocabSet{}, nil)
			},
			expected:    []model.VocabSet{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mocks.NewMockFirestore(ctrl)
			tt.setupMock(mockStore)

			service := &SetService{
				storage: mockStore,
			}

			ctx := userid.ToCtx(context.Background(), userId)
			result, err := service.GetSetList(ctx, userId)

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

func TestSetService_GetSetVocabBatch(t *testing.T) {
	userId := "test-user"
	setId := uuid.New()
	vocabId1 := uuid.New()
	vocabId2 := uuid.New()
	formId1 := uuid.New()
	formId2 := uuid.New()

	vocabSet := &model.VocabSet{
		Id:       setId,
		Name:     "Test Set",
		VocabIds: []uuid.UUID{vocabId1, vocabId2},
	}

	vocabs := []model.Vocab{
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
	}

	tests := []struct {
		name        string
		setId       uuid.UUID
		limit       int
		setupMock   func(*mocks.MockFirestore)
		expected    *model.Batch
		expectError bool
	}{
		{
			name:  "successfully get set vocab batch without limit",
			setId: setId,
			limit: 0,
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					GetVocabSet(gomock.Any(), userId, setId).
					Return(vocabSet, nil)
				mock.EXPECT().
					GetMultipleVocabs(gomock.Any(), userId, vocabSet.VocabIds).
					Return(vocabs, nil)
			},
			expected: &model.Batch{
				Vocabs: vocabs,
			},
			expectError: false,
		},
		{
			name:  "successfully get set vocab batch with limit",
			setId: setId,
			limit: 1,
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					GetVocabSet(gomock.Any(), userId, setId).
					Return(vocabSet, nil)
				mock.EXPECT().
					GetMultipleVocabs(gomock.Any(), userId, vocabSet.VocabIds).
					Return(vocabs, nil)
			},
			expected: &model.Batch{
				Vocabs: vocabs[:1],
			},
			expectError: false,
		},
		{
			name:  "set not found",
			setId: setId,
			limit: 0,
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					GetVocabSet(gomock.Any(), userId, setId).
					Return(nil, nil)
			},
			expected:    nil,
			expectError: true,
		},
		{
			name:  "get set error",
			setId: setId,
			limit: 0,
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					GetVocabSet(gomock.Any(), userId, setId).
					Return(nil, errors.New("get set error"))
			},
			expected:    nil,
			expectError: true,
		},
		{
			name:  "get multiple vocabs error",
			setId: setId,
			limit: 0,
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					GetVocabSet(gomock.Any(), userId, setId).
					Return(vocabSet, nil)
				mock.EXPECT().
					GetMultipleVocabs(gomock.Any(), userId, vocabSet.VocabIds).
					Return(nil, errors.New("get vocabs error"))
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mocks.NewMockFirestore(ctrl)
			tt.setupMock(mockStore)

			service := &SetService{
				storage: mockStore,
			}

			ctx := userid.ToCtx(context.Background(), userId)
			result, err := service.GetSetVocabBatch(ctx, userId, tt.setId, tt.limit)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.limit > 0 {
					require.Len(t, result.Vocabs, tt.limit)
				} else {
					require.Len(t, result.Vocabs, len(vocabs))
				}
			}
		})
	}
}

func TestSetService_AddSet(t *testing.T) {
	userId := "test-user"
	vocabId1 := uuid.New()
	vocabId2 := uuid.New()
	setName := "New Test Set"

	tests := []struct {
		name        string
		setName     string
		vocabIds    []uuid.UUID
		setupMock   func(*mocks.MockFirestore)
		expectError bool
	}{
		{
			name:     "successfully add new set",
			setName:  setName,
			vocabIds: []uuid.UUID{vocabId1, vocabId2},
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					SetVocabSet(gomock.Any(), userId, gomock.Any()).
					DoAndReturn(func(ctx context.Context, userId string, vocabSet model.VocabSet) error {
						require.NotEqual(t, uuid.Nil, vocabSet.Id)
						require.Equal(t, setName, vocabSet.Name)
						require.Equal(t, []uuid.UUID{vocabId1, vocabId2}, vocabSet.VocabIds)
						return nil
					})
			},
			expectError: false,
		},
		{
			name:     "add set with empty vocab list",
			setName:  setName,
			vocabIds: []uuid.UUID{},
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					SetVocabSet(gomock.Any(), userId, gomock.Any()).
					DoAndReturn(func(ctx context.Context, userId string, vocabSet model.VocabSet) error {
						require.NotEqual(t, uuid.Nil, vocabSet.Id)
						require.Equal(t, setName, vocabSet.Name)
						require.Empty(t, vocabSet.VocabIds)
						return nil
					})
			},
			expectError: false,
		},
		{
			name:     "set storage error",
			setName:  setName,
			vocabIds: []uuid.UUID{vocabId1},
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					SetVocabSet(gomock.Any(), userId, gomock.Any()).
					Return(errors.New("storage error"))
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

			service := &SetService{
				storage: mockStore,
			}

			ctx := userid.ToCtx(context.Background(), userId)
			result, err := service.AddSet(ctx, userId, tt.setName, tt.vocabIds)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.NotEqual(t, uuid.Nil, result.Id)
				require.Equal(t, tt.setName, result.Name)
				require.Equal(t, tt.vocabIds, result.VocabIds)
			}
		})
	}
}

func TestSetService_UpdateSet(t *testing.T) {
	userId := "test-user"
	setId := uuid.New()
	vocabId1 := uuid.New()
	vocabId2 := uuid.New()
	updatedName := "Updated Test Set"

	tests := []struct {
		name        string
		setId       uuid.UUID
		setName     string
		vocabIds    []uuid.UUID
		setupMock   func(*mocks.MockFirestore)
		expectError bool
	}{
		{
			name:     "successfully update existing set",
			setId:    setId,
			setName:  updatedName,
			vocabIds: []uuid.UUID{vocabId1, vocabId2},
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					GetVocabSet(gomock.Any(), userId, setId).
					Return(&model.VocabSet{
						Id:       setId,
						Name:     "Old Name",
						VocabIds: []uuid.UUID{},
					}, nil)
				mock.EXPECT().
					SetVocabSet(gomock.Any(), userId, gomock.Any()).
					DoAndReturn(func(ctx context.Context, userId string, vocabSet model.VocabSet) error {
						require.Equal(t, setId, vocabSet.Id)
						require.Equal(t, updatedName, vocabSet.Name)
						require.Equal(t, []uuid.UUID{vocabId1, vocabId2}, vocabSet.VocabIds)
						return nil
					})
			},
			expectError: false,
		},
		{
			name:     "set not found",
			setId:    setId,
			setName:  updatedName,
			vocabIds: []uuid.UUID{vocabId1},
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					GetVocabSet(gomock.Any(), userId, setId).
					Return(nil, nil)
			},
			expectError: true,
		},
		{
			name:     "get set error",
			setId:    setId,
			setName:  updatedName,
			vocabIds: []uuid.UUID{vocabId1},
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					GetVocabSet(gomock.Any(), userId, setId).
					Return(nil, errors.New("get set error"))
			},
			expectError: true,
		},
		{
			name:     "set storage error",
			setId:    setId,
			setName:  updatedName,
			vocabIds: []uuid.UUID{vocabId1},
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					GetVocabSet(gomock.Any(), userId, setId).
					Return(&model.VocabSet{
						Id:       setId,
						Name:     "Old Name",
						VocabIds: []uuid.UUID{},
					}, nil)
				mock.EXPECT().
					SetVocabSet(gomock.Any(), userId, gomock.Any()).
					Return(errors.New("storage error"))
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

			service := &SetService{
				storage: mockStore,
			}

			ctx := userid.ToCtx(context.Background(), userId)
			err := service.UpdateSet(ctx, userId, tt.setId, tt.setName, tt.vocabIds)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSetService_ContextWithoutUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockFirestore(ctrl)
	service := &SetService{
		storage: mockStore,
	}

	ctx := context.Background() // No user ID in context
	userId := "test-user"
	setId := uuid.New()

	// GetSetList will call FetchUserVocabSets
	mockStore.EXPECT().
		FetchUserVocabSets(gomock.Any(), userId).
		Return(nil, errors.New("storage error"))

	_, err := service.GetSetList(ctx, userId)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to fetch user vocab sets")

	// GetSetVocabBatch will call GetVocabSet
	mockStore.EXPECT().
		GetVocabSet(gomock.Any(), userId, setId).
		Return(nil, errors.New("storage error"))

	_, err = service.GetSetVocabBatch(ctx, userId, setId, 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get vocab set")

	// AddSet will call SetVocabSet
	mockStore.EXPECT().
		SetVocabSet(gomock.Any(), userId, gomock.Any()).
		Return(errors.New("storage error"))

	_, err = service.AddSet(ctx, userId, "test", []uuid.UUID{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to add vocab set")

	// UpdateSet will call GetVocabSet
	mockStore.EXPECT().
		GetVocabSet(gomock.Any(), userId, setId).
		Return(nil, errors.New("storage error"))

	err = service.UpdateSet(ctx, userId, setId, "test", []uuid.UUID{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get existing vocab set")

	// RemoveSet will call GetVocabSet
	mockStore.EXPECT().
		GetVocabSet(gomock.Any(), userId, setId).
		Return(nil, errors.New("storage error"))

	err = service.RemoveSet(ctx, userId, setId)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get existing vocab set")
}

func TestSetService_RemoveSet(t *testing.T) {
	userId := "test-user"
	setId := uuid.New()
	vocabId := uuid.New()

	existingSet := &model.VocabSet{
		Id:       setId,
		Name:     "Test Set",
		VocabIds: []uuid.UUID{vocabId},
	}

	tests := []struct {
		name        string
		setId       uuid.UUID
		setupMock   func(*mocks.MockFirestore)
		expectError bool
	}{
		{
			name:  "successfully remove set",
			setId: setId,
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					GetVocabSet(gomock.Any(), userId, setId).
					Return(existingSet, nil)
				mock.EXPECT().
					RemoveVocabSet(gomock.Any(), userId, setId).
					Return(nil)
			},
			expectError: false,
		},
		{
			name:  "set not found",
			setId: setId,
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					GetVocabSet(gomock.Any(), userId, setId).
					Return(nil, nil)
			},
			expectError: true,
		},
		{
			name:  "get set error",
			setId: setId,
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					GetVocabSet(gomock.Any(), userId, setId).
					Return(nil, errors.New("get set error"))
			},
			expectError: true,
		},
		{
			name:  "remove set storage error",
			setId: setId,
			setupMock: func(mock *mocks.MockFirestore) {
				mock.EXPECT().
					GetVocabSet(gomock.Any(), userId, setId).
					Return(existingSet, nil)
				mock.EXPECT().
					RemoveVocabSet(gomock.Any(), userId, setId).
					Return(errors.New("storage error"))
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

			service := &SetService{
				storage: mockStore,
			}

			ctx := userid.ToCtx(context.Background(), userId)
			err := service.RemoveSet(ctx, userId, tt.setId)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
