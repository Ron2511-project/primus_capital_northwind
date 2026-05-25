package httpadapter

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/ron/northwind-collections/internal/application"
)

func NewRouter(svc *application.CollectionService) http.Handler {
	h := NewHandler(svc)
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	origin := os.Getenv("CORS_ORIGIN")
	if origin == "" {
		origin = "http://localhost:5173"
	}
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{origin},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Get("/health", h.Health)
	r.Get("/api/docs", serveOpenAPI)
	r.Get("/api/docs/openapi.yaml", serveOpenAPIFile)

	r.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})
	r.Get("/swagger/index.html", serveSwaggerUI)

	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/dashboard/summary", h.GetSummary)
		api.Get("/collection-queue", h.GetQueue)
		api.Get("/customers/{id}", h.GetCustomer)
		api.Post("/customers/{id}/actions", h.CreateAction)
	})

	return r
}

func serveOpenAPI(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/api/docs/openapi.yaml", http.StatusTemporaryRedirect)
}

func serveOpenAPIFile(w http.ResponseWriter, r *http.Request) {
	body, err := os.ReadFile("docs/openapi.yaml")
	if err != nil {
		http.Error(w, "openapi not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/yaml")
	w.Write(body)
}
