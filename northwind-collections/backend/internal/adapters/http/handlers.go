package httpadapter

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ron/northwind-collections/internal/application"
	"github.com/ron/northwind-collections/internal/domain"
)

type Handler struct {
	svc *application.CollectionService
}

func NewHandler(svc *application.CollectionService) *Handler {
	return &Handler{svc: svc}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) GetSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.svc.GetSummary(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load summary")
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (h *Handler) GetQueue(w http.ResponseWriter, r *http.Request) {
	var segment *domain.Segment
	if s := r.URL.Query().Get("segment"); s != "" {
		seg := domain.Segment(s)
		segment = &seg
	}
	var priority *domain.PriorityLevel
	if p := r.URL.Query().Get("priority"); p != "" {
		pr := domain.PriorityLevel(p)
		priority = &pr
	}
	queue, err := h.svc.GetQueue(r.Context(), segment, priority)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load queue")
		return
	}
	writeJSON(w, http.StatusOK, queue)
}

func (h *Handler) GetCustomer(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid customer id")
		return
	}
	detail, err := h.svc.GetCustomerDetail(r.Context(), id)
	if err != nil {
		if errors.Is(err, application.ErrCustomerNotFound) {
			writeError(w, http.StatusNotFound, "customer not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load customer")
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

type createActionRequest struct {
	ActionType string `json:"action_type"`
	Notes      string `json:"notes"`
	CreatedBy  string `json:"created_by"`
}

func (h *Handler) CreateAction(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid customer id")
		return
	}
	var req createActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	req.ActionType = strings.TrimSpace(req.ActionType)
	req.Notes = strings.TrimSpace(req.Notes)
	if req.ActionType == "" {
		writeError(w, http.StatusBadRequest, "action_type is required")
		return
	}

	action, err := h.svc.CreateAction(r.Context(), application.CreateActionInput{
		CustomerID: id,
		ActionType: domain.ActionType(req.ActionType),
		Notes:      req.Notes,
		CreatedBy:  req.CreatedBy,
	})
	if err != nil {
		if errors.Is(err, application.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, application.ErrCustomerNotFound) {
			writeError(w, http.StatusNotFound, "customer not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create action")
		return
	}
	writeJSON(w, http.StatusCreated, action)
}
