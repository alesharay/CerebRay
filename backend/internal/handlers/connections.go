package handlers

import (
	"encoding/json"
	"net/http"

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
	noteID, err := URLParamInt64(r, "id")
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid note ID")
		return
	}

	connections, err := h.queries.ListConnectionsForNote(r.Context(), noteID)
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
	var req createConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	conn, err := h.queries.CreateConnection(r.Context(), sqlc.CreateConnectionParams{
		SourceID: req.SourceID,
		TargetID: req.TargetID,
		Label:    req.Label,
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to create connection")
		return
	}
	JSON(w, http.StatusCreated, conn)
}

func (h *ConnectionHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	connID, err := URLParamInt64(r, "id")
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid connection ID")
		return
	}

	if err := h.queries.DeleteConnection(r.Context(), connID); err != nil {
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
