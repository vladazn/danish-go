package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/vladazn/danish/app/model"
)

// @Summary Add vocab
// @Tags vocab
// @Accept json
// @Produce json
// @Param vocab body model.Vocab true "Vocabulary"
// @Success 200 {object} model.Vocab
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /vocab [post]
func (h *handler) handleAddWord(w http.ResponseWriter, r *http.Request) {
	var v model.Vocab
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	addedWord, err := h.dict.AddWord(r.Context(), v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(addedWord)
}

// @Summary Update vocab
// @Tags vocab
// @Accept json
// @Produce json
// @Param vocab body model.Vocab true "Vocabulary"
// @Success 200 {object} model.Vocab
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /vocab [put]
func (h *handler) handleUpdateWord(w http.ResponseWriter, r *http.Request) {
	var v model.Vocab
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updatedWord, err := h.dict.UpdateWord(r.Context(), v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedWord)
}

// @Summary Remove vocab
// @Tags vocab
// @Param id path string true "Vocab ID"
// @Success 200
// @Failure 400,500
// @Router /vocab/{id} [delete]
func (h *handler) handleRemoveWord(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	vocabId, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid UUID", http.StatusBadRequest)
		return
	}

	if err := h.dict.RemoveWord(r.Context(), vocabId); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary Get all vocab for user
// @Tags vocab
// @Produce json
// @Success 200 {array} model.Vocab
// @Failure 500
// @Router /vocab [get]
func (h *handler) handleGetAllWords(w http.ResponseWriter, r *http.Request) {
	vocabs, err := h.dict.GetAllWords(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vocabs)
}
