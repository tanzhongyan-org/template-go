package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tanzhongyan-org/template-go/internal/db"
)

type NoteResponse struct {
	ID        string     `json:"id"`
	OwnerID   string     `json:"owner_id"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type CreateNoteRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type UpdateNoteRequest struct {
	Title *string `json:"title,omitempty"`
	Body  *string `json:"body,omitempty"`
}

func toNoteResponse(n db.Note) NoteResponse {
	var updatedAt *time.Time
	if n.UpdatedAt.Valid {
		t := n.UpdatedAt.Time
		updatedAt = &t
	}
	var createdAt time.Time
	if n.CreatedAt.Valid {
		createdAt = n.CreatedAt.Time
	} else {
		createdAt = time.Now().UTC()
	}

	return NoteResponse{
		ID:        n.ID,
		OwnerID:   n.OwnerID,
		Title:     n.Title,
		Body:      n.Body,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

func extractOwnerID(r *http.Request) string {
	owner := r.Header.Get("x-user-id")
	if strings.TrimSpace(owner) == "" {
		return "test-owner"
	}
	return strings.TrimSpace(owner)
}

type NotesHandler struct {
	queries *db.Queries
}

func NewNotesHandler(queries *db.Queries) *NotesHandler {
	return &NotesHandler{queries: queries}
}

func (h *NotesHandler) List(w http.ResponseWriter, r *http.Request) {
	ownerID := extractOwnerID(r)
	notes, err := h.queries.ListNotes(r.Context(), ownerID)
	if err != nil {
		http.Error(w, `{"error":{"message":"Failed to list notes"}}`, http.StatusInternalServerError)
		return
	}

	res := make([]NoteResponse, 0, len(notes))
	for _, n := range notes {
		res = append(res, toNoteResponse(n))
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func (h *NotesHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ownerID := extractOwnerID(r)

	note, err := h.queries.GetNote(r.Context(), db.GetNoteParams{
		ID:      id,
		OwnerID: ownerID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, `{"error":{"message":"Note not found"}}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":{"message":"Failed to get note"}}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(toNoteResponse(note))
}

func (h *NotesHandler) Create(w http.ResponseWriter, r *http.Request) {
	ownerID := extractOwnerID(r)

	var req CreateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":{"message":"Invalid JSON payload"}}`, http.StatusBadRequest)
		return
	}

	trimmedTitle := strings.TrimSpace(req.Title)
	if trimmedTitle == "" {
		http.Error(w, `{"error":{"message":"Title cannot be empty"}}`, http.StatusBadRequest)
		return
	}
	if len(trimmedTitle) > 200 {
		http.Error(w, `{"error":{"message":"Title must not exceed 200 characters"}}`, http.StatusBadRequest)
		return
	}

	trimmedBody := strings.TrimSpace(req.Body)
	if trimmedBody == "" {
		http.Error(w, `{"error":{"message":"Body cannot be empty"}}`, http.StatusBadRequest)
		return
	}

	id := uuid.New().String()
	note, err := h.queries.CreateNote(r.Context(), db.CreateNoteParams{
		ID:      id,
		OwnerID: ownerID,
		Title:   trimmedTitle,
		Body:    trimmedBody,
	})
	if err != nil {
		http.Error(w, `{"error":{"message":"Failed to create note"}}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(toNoteResponse(note))
}

func (h *NotesHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ownerID := extractOwnerID(r)

	var req UpdateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":{"message":"Invalid JSON payload"}}`, http.StatusBadRequest)
		return
	}

	var titleParam pgtype.Text
	if req.Title != nil {
		t := strings.TrimSpace(*req.Title)
		if t == "" {
			http.Error(w, `{"error":{"message":"Title cannot be empty"}}`, http.StatusBadRequest)
			return
		}
		if len(t) > 200 {
			http.Error(w, `{"error":{"message":"Title must not exceed 200 characters"}}`, http.StatusBadRequest)
			return
		}
		titleParam = pgtype.Text{String: t, Valid: true}
	}

	var bodyParam pgtype.Text
	if req.Body != nil {
		b := strings.TrimSpace(*req.Body)
		if b == "" {
			http.Error(w, `{"error":{"message":"Body cannot be empty"}}`, http.StatusBadRequest)
			return
		}
		bodyParam = pgtype.Text{String: b, Valid: true}
	}

	note, err := h.queries.UpdateNote(r.Context(), db.UpdateNoteParams{
		ID:      id,
		OwnerID: ownerID,
		Title:   titleParam,
		Body:    bodyParam,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, `{"error":{"message":"Note not found"}}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":{"message":"Failed to update note"}}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(toNoteResponse(note))
}

func (h *NotesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ownerID := extractOwnerID(r)

	rows, err := h.queries.DeleteNote(r.Context(), db.DeleteNoteParams{
		ID:      id,
		OwnerID: ownerID,
	})
	if err != nil {
		http.Error(w, `{"error":{"message":"Failed to delete note"}}`, http.StatusInternalServerError)
		return
	}
	if rows == 0 {
		http.Error(w, `{"error":{"message":"Note not found"}}`, http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}