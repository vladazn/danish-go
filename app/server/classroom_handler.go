package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/vladazn/danish/app/classroom"
	"github.com/vladazn/danish/app/model"
	"github.com/vladazn/danish/common/userid"
)

type ClassroomHandlerParams struct {
	fx.In
	FirebaseApp *FirebaseClient
	Logger      *zap.Logger
	Dict        *classroom.Dictionary
	Pool        *classroom.WordPool
	Set         *classroom.SetService
}

// @Summary Get a new batch of words from the pool
// @Description Fetches a batch of 20 vocab entries from the user's learning pool
// @Tags pool
// @Produce json
// @Success 200 {object} model.Batch
// @Failure 500 {string} string "Server error"
// @Router /classroom/batch [get]
func (h *handler) handleGetBatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId := userid.MustFromCtx(ctx) // replace with your actual userID extraction

	batch, err := h.pool.GetBatch(ctx, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(batch)
}

type BatchResult struct {
	WithoutMistake []model.Vocab `json:"without_mistake"`
	WithMistake    []model.Vocab `json:"with_mistake"`
}

// @Summary Remove vocab from the learning pool
// @Description Removes one or more vocab items from the user’s current pool
// @Tags pool
// @Accept json
// @Produce json
// @Param vocab body BatchResult true "Vocab items to remove"
// @Success 200
// @Failure 400 {string} string "Invalid request"
// @Failure 500 {string} string "Server error"
// @Router /classroom/batch [post]
func (h *handler) handleBatchResult(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId := userid.MustFromCtx(ctx) // replace with your actual userID extraction

	var batchResult BatchResult
	if err := json.NewDecoder(r.Body).Decode(&batchResult); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	b, _ := json.Marshal(batchResult)
	println(string(b))

	if err := h.pool.RemoveFromPool(ctx, userId, batchResult.WithoutMistake); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.dict.RegisterProgress(ctx, batchResult.WithoutMistake, batchResult.WithMistake); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary Get list of vocab sets
// @Description Fetches all vocab sets for the authenticated user
// @Tags sets
// @Produce json
// @Success 200 {array} model.VocabSet
// @Failure 500 {string} string "Server error"
// @Router /classroom/sets [get]
func (h *handler) handleGetSetList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId := userid.MustFromCtx(ctx)

	sets, err := h.set.GetSetList(ctx, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(sets) == 0 {
		sets = []model.VocabSet{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sets)
}

// @Summary Get vocab batch from a specific set
// @Description Fetches a batch of vocabulary items from a specific vocab set
// @Tags sets
// @Produce json
// @Param setId path string true "Set ID"
// @Param limit query int false "Limit number of items (default: all)"
// @Success 200 {object} model.Batch
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Set not found"
// @Failure 500 {string} string "Server error"
// @Router /classroom/sets/{setId}/batch [get]
func (h *handler) handleGetSetVocabBatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId := userid.MustFromCtx(ctx)

	// Extract setId from URL path
	setIdStr := chi.URLParam(r, "setId")
	setId, err := uuid.Parse(setIdStr)
	if err != nil {
		http.Error(w, "Invalid set ID", http.StatusBadRequest)
		return
	}

	// Parse limit query parameter
	limit := 0 // 0 means no limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	batch, err := h.set.GetSetVocabBatch(ctx, userId, setId, limit)
	if err != nil {
		if err.Error() == "vocab set not found" {
			http.Error(w, "Set not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(batch)
}

// @Summary Add a new vocab set
// @Description Creates a new vocab set for the authenticated user
// @Tags sets
// @Accept json
// @Produce json
// @Param set body AddSetRequest true "Set information"
// @Success 200 {object} model.VocabSet
// @Failure 400 {string} string "Invalid request"
// @Failure 500 {string} string "Server error"
// @Router /classroom/sets [post]
func (h *handler) handleAddSet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId := userid.MustFromCtx(ctx)

	var req AddSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Set name is required", http.StatusBadRequest)
		return
	}

	set, err := h.set.AddSet(ctx, userId, req.Name, req.VocabIds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(set)
}

// @Summary Update a vocab set
// @Description Updates an existing vocab set for the authenticated user
// @Tags sets
// @Accept json
// @Produce json
// @Param setId path string true "Set ID"
// @Param set body UpdateSetRequest true "Updated set information"
// @Success 200
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Set not found"
// @Failure 500 {string} string "Server error"
// @Router /classroom/sets/{setId} [put]
func (h *handler) handleUpdateSet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId := userid.MustFromCtx(ctx)

	// Extract setId from URL path
	setIdStr := chi.URLParam(r, "setId")
	setId, err := uuid.Parse(setIdStr)
	if err != nil {
		http.Error(w, "Invalid set ID", http.StatusBadRequest)
		return
	}

	var req UpdateSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Set name is required", http.StatusBadRequest)
		return
	}

	err = h.set.UpdateSet(ctx, userId, setId, req.Name, req.VocabIds)
	if err != nil {
		if err.Error() == "vocab set not found" {
			http.Error(w, "Set not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary Remove a vocab set
// @Description Removes a vocab set for the authenticated user
// @Tags sets
// @Produce json
// @Param setId path string true "Set ID"
// @Success 200
// @Failure 400 {string} string "Invalid set ID"
// @Failure 404 {string} string "Set not found"
// @Failure 500 {string} string "Server error"
// @Router /classroom/sets/{setId} [delete]
func (h *handler) handleRemoveSet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId := userid.MustFromCtx(ctx)

	// Extract setId from URL path
	setIdStr := chi.URLParam(r, "setId")
	setId, err := uuid.Parse(setIdStr)
	if err != nil {
		http.Error(w, "Invalid set ID", http.StatusBadRequest)
		return
	}

	err = h.set.RemoveSet(ctx, userId, setId)
	if err != nil {
		if err.Error() == "vocab set not found" {
			http.Error(w, "Set not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Request/Response types
type AddSetRequest struct {
	Name     string      `json:"name"`
	VocabIds []uuid.UUID `json:"vocab_ids"`
}

type UpdateSetRequest struct {
	Name     string      `json:"name"`
	VocabIds []uuid.UUID `json:"vocab_ids"`
}
