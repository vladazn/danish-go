package server

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/vladazn/danish/app/classroom"
)

type RouterParams struct {
	fx.In
	FirebaseApp *FirebaseClient
	Logger      *zap.Logger
	Dict        *classroom.Dictionary
	Pool        *classroom.WordPool
	Set         *classroom.SetService
}

func NewRouter(p RouterParams) *chi.Mux {
	r := chi.NewRouter()
	log := p.Logger.With(zap.String("type", "http"))
	r.Use(cors.Handler(cors.Options{
		AllowOriginFunc: func(r *http.Request, origin string) bool {
			// Allow any origin that starts with "http://localhost"
			return strings.HasPrefix(origin, "https://danish-edu.web.app") ||
				strings.HasPrefix(origin, "https://danish-edu.firebaseapp.com") ||
				origin == "http://localhost" ||
				len(origin) > 16 && origin[:17] == "http://localhost:"
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(loggerMiddleware(log))
	r.Use(middleware.Recoverer)
	r.Use(firebaseAuthMiddleware(p.FirebaseApp, log))

	h := &handler{p.Dict, p.Pool, p.Set}

	r.Route("/vocab", func(r chi.Router) {
		r.Post("/", h.handleAddWord)
		r.Put("/", h.handleUpdateWord)
		r.Delete("/{id}", h.handleRemoveWord)
		r.Get("/", h.handleGetAllWords)
	})

	r.Route("/classroom", func(r chi.Router) {
		r.Get("/batch", h.handleGetBatch)
		r.Post("/batch", h.handleBatchResult)

		r.Route("/sets", func(r chi.Router) {
			r.Get("/", h.handleGetSetList)
			r.Post("/", h.handleAddSet)
			r.Get("/{setId}/batch", h.handleGetSetVocabBatch)
			r.Put("/{setId}", h.handleUpdateSet)
			r.Delete("/{setId}", h.handleRemoveSet)
		})
	})

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return r
}
