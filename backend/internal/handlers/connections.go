package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/aray/cerebray/backend/db/sqlc"
	"github.com/aray/cerebray/backend/internal/middleware"
)

type ConnectionHandlers struct {
	queries sqlc.Querier
}

func NewConnectionHandlers(q sqlc.Querier) *ConnectionHandlers {
	return &ConnectionHandlers{queries: q}
}

func (h *ConnectionHandlers) ListForNote(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	noteID, err := URLParamInt64(r, "id")
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid note ID")
		return
	}

	connections, err := h.queries.ListConnectionsForNote(r.Context(), sqlc.ListConnectionsForNoteParams{
		SourceID: noteID,
		UserID:   userID,
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to list connections")
		return
	}
	JSON(w, http.StatusOK, connections)
}

type createConnectionRequest struct {
	SourceID int64  `json:"source_id"`
	TargetID int64  `json:"target_id"`
	Label    string `json:"label"`
}

func (h *ConnectionHandlers) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req createConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	conn, err := h.queries.CreateConnection(r.Context(), sqlc.CreateConnectionParams{
		SourceID: req.SourceID,
		TargetID: req.TargetID,
		Label:    req.Label,
		UserID:   userID,
	})
	if err != nil {
		// The insert is guarded by an EXISTS on both notes, so no rows means
		// at least one of them is missing or belongs to someone else.
		if errors.Is(err, pgx.ErrNoRows) {
			Error(w, http.StatusNotFound, "note not found")
			return
		}
		// UNIQUE (source_id, target_id) and CHECK (source_id <> target_id) are
		// both reachable by ordinary use: re-linking two notes, or linking a
		// note to itself. Neither is a server error.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				Error(w, http.StatusConflict, "these notes are already connected")
				return
			case "23514":
				Error(w, http.StatusBadRequest, "a note cannot be connected to itself")
				return
			}
		}
		Error(w, http.StatusInternalServerError, "failed to create connection")
		return
	}
	JSON(w, http.StatusCreated, conn)
}

func (h *ConnectionHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	connID, err := URLParamInt64(r, "id")
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid connection ID")
		return
	}

	if err := h.queries.DeleteConnection(r.Context(), sqlc.DeleteConnectionParams{
		ID:     connID,
		UserID: userID,
	}); err != nil {
		Error(w, http.StatusInternalServerError, "failed to delete connection")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type graphNode struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Type  string `json:"type"`
}

type graphEdge struct {
	Source int64  `json:"source"`
	Target int64  `json:"target"`
	Label  string `json:"label"`
}

func (h *ConnectionHandlers) GraphData(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	data, err := h.queries.GetGraphData(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to get graph data")
		return
	}

	nodeMap := map[int64]graphNode{}
	edges := make([]graphEdge, 0, len(data))

	for _, row := range data {
		nodeMap[row.SourceID] = graphNode{ID: row.SourceID, Title: row.SourceTitle, Type: string(row.SourceType)}
		nodeMap[row.TargetID] = graphNode{ID: row.TargetID, Title: row.TargetTitle, Type: string(row.TargetType)}
		edges = append(edges, graphEdge{Source: row.SourceID, Target: row.TargetID, Label: row.Label})
	}

	nodes := make([]graphNode, 0, len(nodeMap))
	for _, n := range nodeMap {
		nodes = append(nodes, n)
	}

	JSON(w, http.StatusOK, map[string]any{
		"nodes": nodes,
		"edges": edges,
	})
}
