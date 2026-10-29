package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tanzhongyan-org/template-go/internal/auth"
	"github.com/tanzhongyan-org/template-go/internal/config"
	"github.com/tanzhongyan-org/template-go/internal/db"
)

// maxTitleLength mirrors VARCHAR(200) in sql/schema.sql.
const maxTitleLength = 200

// TEMPLATE-TODO(example): replace the notes slice with your own domain.
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

type ownerKey struct{}

// ownerFrom returns the caller identity stored by requireOwner.
func ownerFrom(ctx context.Context) string {
	owner, _ := ctx.Value(ownerKey{}).(string)
	return owner
}

type NotesHandler struct {
	queries *db.Queries
	trust   bool

	jwksURL  string
	issuer   string
	audience string

	mu       sync.Mutex
	verifier *auth.Verifier
}

func NewNotesHandler(queries *db.Queries, cfg *config.Config) *NotesHandler {
	return &NotesHandler{
		queries:  queries,
		trust:    cfg.TrustIdentityHeader,
		jwksURL:  cfg.JWKSURL,
		issuer:   cfg.JWTIssuer,
		audience: cfg.JWTAudience,
	}
}

// Routes returns the notes routes behind the identity check.
func (h *NotesHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(h.requireOwner)
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.Get)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	return r
}

// requireOwner is the single identity boundary. With AUTH_JWKS_URL set it
// verifies the Authorization header against the product JWKS; otherwise it
// honours x-user-id only when TRUST_IDENTITY_HEADER=true. It fails closed:
// any other shape answers 401.
// TEMPLATE-TODO(auth): replace header identity with real authentication
func (h *NotesHandler) requireOwner(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.jwksURL != "" {
			h.requireJWT(w, r, next)
			return
		}
		owner := strings.TrimSpace(r.Header.Get("x-user-id"))
		if !h.trust || owner == "" {
			writeError(w, http.StatusUnauthorized, "Authentication required")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ownerKey{}, owner)))
	})
}

// requireJWT verifies the Authorization header against the product JWKS and
// stores the subject as the caller identity.
func (h *NotesHandler) requireJWT(w http.ResponseWriter, r *http.Request, next http.Handler) {
	raw, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok || strings.TrimSpace(raw) == "" {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	v, err := h.verifierFor()
	if err != nil {
		log.Printf("auth: %v", err)
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	subject, err := v.Verify(strings.TrimSpace(raw))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ownerKey{}, subject)))
}

// verifierFor lazily builds the JWT verifier on first use and caches it, so
// the JWKS document is fetched only when the JWT path serves traffic.
func (h *NotesHandler) verifierFor() (*auth.Verifier, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.verifier != nil {
		return h.verifier, nil
	}
	v, err := auth.NewVerifier(h.jwksURL, h.issuer, h.audience)
	if err != nil {
		return nil, err
	}
	h.verifier = v
	return v, nil
}

// cleanTitle trims and validates a note title. A non-empty message is the 400 reason.
func cleanTitle(raw string) (string, string) {
	t := strings.TrimSpace(raw)
	if t == "" {
		return "", "Title cannot be empty"
	}
	if len(t) > maxTitleLength {
		return "", "Title must not exceed 200 characters"
	}
	return t, ""
}

// cleanBody trims and validates a note body. A non-empty message is the 400 reason.
func cleanBody(raw string) (string, string) {
	b := strings.TrimSpace(raw)
	if b == "" {
		return "", "Body cannot be empty"
	}
	return b, ""
}

func (h *NotesHandler) List(w http.ResponseWriter, r *http.Request) {
	notes, err := h.queries.ListNotes(r.Context(), ownerFrom(r.Context()))
	if err != nil {
		log.Printf("list notes: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to list notes")
		return
	}

	res := make([]NoteResponse, 0, len(notes))
	for _, n := range notes {
		res = append(res, toNoteResponse(n))
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *NotesHandler) Get(w http.ResponseWriter, r *http.Request) {
	note, err := h.queries.GetNote(r.Context(), db.GetNoteParams{
		ID:      chi.URLParam(r, "id"),
		OwnerID: ownerFrom(r.Context()),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Note not found")
			return
		}
		log.Printf("get note: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to get note")
		return
	}
	writeJSON(w, http.StatusOK, toNoteResponse(note))
}

func (h *NotesHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateNoteRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	title, msg := cleanTitle(req.Title)
	if msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	body, msg := cleanBody(req.Body)
	if msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	note, err := h.queries.CreateNote(r.Context(), db.CreateNoteParams{
		ID:      uuid.New().String(),
		OwnerID: ownerFrom(r.Context()),
		Title:   title,
		Body:    body,
	})
	if err != nil {
		log.Printf("create note: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to create note")
		return
	}
	writeJSON(w, http.StatusCreated, toNoteResponse(note))
}

func (h *NotesHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req UpdateNoteRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	var titleParam pgtype.Text
	if req.Title != nil {
		title, msg := cleanTitle(*req.Title)
		if msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}
		titleParam = pgtype.Text{String: title, Valid: true}
	}

	var bodyParam pgtype.Text
	if req.Body != nil {
		body, msg := cleanBody(*req.Body)
		if msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}
		bodyParam = pgtype.Text{String: body, Valid: true}
	}

	note, err := h.queries.UpdateNote(r.Context(), db.UpdateNoteParams{
		ID:      chi.URLParam(r, "id"),
		OwnerID: ownerFrom(r.Context()),
		Title:   titleParam,
		Body:    bodyParam,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Note not found")
			return
		}
		log.Printf("update note: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to update note")
		return
	}
	writeJSON(w, http.StatusOK, toNoteResponse(note))
}

func (h *NotesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	rows, err := h.queries.DeleteNote(r.Context(), db.DeleteNoteParams{
		ID:      chi.URLParam(r, "id"),
		OwnerID: ownerFrom(r.Context()),
	})
	if err != nil {
		log.Printf("delete note: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to delete note")
		return
	}
	if rows == 0 {
		writeError(w, http.StatusNotFound, "Note not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
