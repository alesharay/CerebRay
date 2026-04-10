package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aray/cerebray/backend/db/sqlc"
)

func TestNoteHandlers_List(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		setup      func(q *fakeQuerier)
		wantStatus int
	}{
		{
			name:  "list all notes for user",
			query: "",
			setup: func(q *fakeQuerier) {
				q.ListNotesByUserFn = func(_ context.Context, arg sqlc.ListNotesByUserParams) ([]sqlc.ListNotesByUserRow, error) {
					if arg.UserID != 1 {
						t.Errorf("expected userID 1, got %d", arg.UserID)
					}
					if arg.Limit != 50 {
						t.Errorf("expected default limit 50, got %d", arg.Limit)
					}
					return []sqlc.ListNotesByUserRow{
						{ID: 10, UserID: 1, Title: "Note A"},
						{ID: 11, UserID: 1, Title: "Note B"},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "filter by status",
			query: "?status=active",
			setup: func(q *fakeQuerier) {
				q.ListNotesByStatusFn = func(_ context.Context, arg sqlc.ListNotesByStatusParams) ([]sqlc.Note, error) {
					if arg.Status != sqlc.NoteStatusActive {
						t.Errorf("expected status active, got %s", arg.Status)
					}
					return []sqlc.Note{sampleNote(10, 1)}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "filter by tag",
			query: "?tag=golang",
			setup: func(q *fakeQuerier) {
				q.ListNotesByTagFn = func(_ context.Context, arg sqlc.ListNotesByTagParams) ([]sqlc.Note, error) {
					if arg.Name != "golang" {
						t.Errorf("expected tag golang, got %s", arg.Name)
					}
					return []sqlc.Note{sampleNote(10, 1)}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "custom limit and offset",
			query: "?limit=10&offset=5",
			setup: func(q *fakeQuerier) {
				q.ListNotesByUserFn = func(_ context.Context, arg sqlc.ListNotesByUserParams) ([]sqlc.ListNotesByUserRow, error) {
					if arg.Limit != 10 {
						t.Errorf("expected limit 10, got %d", arg.Limit)
					}
					if arg.Offset != 5 {
						t.Errorf("expected offset 5, got %d", arg.Offset)
					}
					return nil, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "db error returns 500",
			query: "",
			setup: func(q *fakeQuerier) {
				q.ListNotesByUserFn = func(_ context.Context, _ sqlc.ListNotesByUserParams) ([]sqlc.ListNotesByUserRow, error) {
					return nil, fmt.Errorf("db down")
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &fakeQuerier{}
			tt.setup(q)
			h := NewNoteHandlers(q)

			r := reqWithUser("GET", "/api/v1/notes"+tt.query, nil, 1)
			w := httptest.NewRecorder()

			h.List(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("List() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestNoteHandlers_Create(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setup      func(q *fakeQuerier)
		wantStatus int
	}{
		{
			name: "valid note",
			body: `{"title":"My Note","body":"Some content"}`,
			setup: func(q *fakeQuerier) {
				q.CreateNoteFn = func(_ context.Context, arg sqlc.CreateNoteParams) (sqlc.Note, error) {
					if arg.Title != "My Note" {
						t.Errorf("expected title My Note, got %s", arg.Title)
					}
					if arg.NoteType != sqlc.NoteTypeConcept {
						t.Errorf("expected default note_type concept, got %s", arg.NoteType)
					}
					if arg.Status != sqlc.NoteStatusFleeting {
						t.Errorf("expected default status fleeting, got %s", arg.Status)
					}
					return sampleNote(1, arg.UserID), nil
				}
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid JSON",
			body:       `{bad json`,
			setup:      func(q *fakeQuerier) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing title",
			body:       `{"body":"No title here"}`,
			setup:      func(q *fakeQuerier) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "db error returns 500",
			body: `{"title":"Fail Note"}`,
			setup: func(q *fakeQuerier) {
				q.CreateNoteFn = func(_ context.Context, _ sqlc.CreateNoteParams) (sqlc.Note, error) {
					return sqlc.Note{}, fmt.Errorf("insert failed")
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &fakeQuerier{}
			tt.setup(q)
			h := NewNoteHandlers(q)

			r := reqWithUser("POST", "/api/v1/notes", strings.NewReader(tt.body), 1)
			w := httptest.NewRecorder()

			h.Create(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("Create() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestNoteHandlers_Get(t *testing.T) {
	tests := []struct {
		name       string
		noteID     string
		setup      func(q *fakeQuerier)
		wantStatus int
	}{
		{
			name:   "existing note",
			noteID: "42",
			setup: func(q *fakeQuerier) {
				q.GetNoteByIDFn = func(_ context.Context, arg sqlc.GetNoteByIDParams) (sqlc.Note, error) {
					if arg.ID != 42 {
						t.Errorf("expected note ID 42, got %d", arg.ID)
					}
					return sampleNote(42, 1), nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "not found",
			noteID: "999",
			setup: func(q *fakeQuerier) {
				q.GetNoteByIDFn = func(_ context.Context, _ sqlc.GetNoteByIDParams) (sqlc.Note, error) {
					return sqlc.Note{}, fmt.Errorf("no rows")
				}
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid ID",
			noteID:     "abc",
			setup:      func(q *fakeQuerier) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &fakeQuerier{}
			tt.setup(q)
			h := NewNoteHandlers(q)

			r := reqWithUserAndChiCtx("GET", "/api/v1/notes/"+tt.noteID, nil, 1, map[string]string{"id": tt.noteID})
			w := httptest.NewRecorder()

			h.Get(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("Get() status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var note sqlc.Note
				decodeJSON(t, w.Body, &note)
				if note.ID != 42 {
					t.Errorf("expected note ID 42, got %d", note.ID)
				}
			}
		})
	}
}

func TestNoteHandlers_Update(t *testing.T) {
	tests := []struct {
		name       string
		noteID     string
		body       string
		setup      func(q *fakeQuerier)
		wantStatus int
	}{
		{
			name:   "successful update",
			noteID: "10",
			body:   `{"title":"Updated Title","body":"New body","note_type":"concept","status":"active","tlp":"clear"}`,
			setup: func(q *fakeQuerier) {
				q.UpdateNoteFn = func(_ context.Context, arg sqlc.UpdateNoteParams) (sqlc.Note, error) {
					if arg.ID != 10 {
						t.Errorf("expected note ID 10, got %d", arg.ID)
					}
					if arg.Title != "Updated Title" {
						t.Errorf("expected title Updated Title, got %s", arg.Title)
					}
					n := sampleNote(10, 1)
					n.Title = arg.Title
					return n, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid JSON body",
			noteID:     "10",
			body:       `not json`,
			setup:      func(q *fakeQuerier) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid note ID",
			noteID:     "xyz",
			body:       `{"title":"X"}`,
			setup:      func(q *fakeQuerier) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "db error returns 500",
			noteID: "10",
			body:   `{"title":"Fail"}`,
			setup: func(q *fakeQuerier) {
				q.UpdateNoteFn = func(_ context.Context, _ sqlc.UpdateNoteParams) (sqlc.Note, error) {
					return sqlc.Note{}, fmt.Errorf("update failed")
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &fakeQuerier{}
			tt.setup(q)
			h := NewNoteHandlers(q)

			r := reqWithUserAndChiCtx("PUT", "/api/v1/notes/"+tt.noteID, strings.NewReader(tt.body), 1, map[string]string{"id": tt.noteID})
			w := httptest.NewRecorder()

			h.Update(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("Update() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestNoteHandlers_Delete(t *testing.T) {
	tests := []struct {
		name       string
		noteID     string
		setup      func(q *fakeQuerier)
		wantStatus int
	}{
		{
			name:   "successful delete",
			noteID: "10",
			setup: func(q *fakeQuerier) {
				q.DeleteNoteFn = func(_ context.Context, arg sqlc.DeleteNoteParams) error {
					if arg.ID != 10 || arg.UserID != 1 {
						t.Errorf("unexpected delete params: ID=%d UserID=%d", arg.ID, arg.UserID)
					}
					return nil
				}
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "invalid ID",
			noteID:     "nope",
			setup:      func(q *fakeQuerier) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "db error returns 500",
			noteID: "10",
			setup: func(q *fakeQuerier) {
				q.DeleteNoteFn = func(_ context.Context, _ sqlc.DeleteNoteParams) error {
					return fmt.Errorf("delete failed")
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &fakeQuerier{}
			tt.setup(q)
			h := NewNoteHandlers(q)

			r := reqWithUserAndChiCtx("DELETE", "/api/v1/notes/"+tt.noteID, nil, 1, map[string]string{"id": tt.noteID})
			w := httptest.NewRecorder()

			h.Delete(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("Delete() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestNoteHandlers_Promote(t *testing.T) {
	tests := []struct {
		name       string
		noteID     string
		setup      func(q *fakeQuerier)
		wantStatus int
	}{
		{
			name:   "successful promote",
			noteID: "5",
			setup: func(q *fakeQuerier) {
				q.GetNoteByIDFn = func(_ context.Context, arg sqlc.GetNoteByIDParams) (sqlc.Note, error) {
					return sampleNote(5, 1), nil
				}
				q.CreateConversationFn = func(_ context.Context, arg sqlc.CreateConversationParams) (sqlc.Conversation, error) {
					return sqlc.Conversation{ID: 100, UserID: 1, Title: arg.Title}, nil
				}
				q.UpdateNoteSourceChatFn = func(_ context.Context, _ sqlc.UpdateNoteSourceChatParams) (sqlc.Note, error) {
					return sampleNote(5, 1), nil
				}
				q.UpdateNoteStatusFn = func(_ context.Context, arg sqlc.UpdateNoteStatusParams) (sqlc.Note, error) {
					if arg.Status != sqlc.NoteStatusActive {
						t.Errorf("expected status active, got %s", arg.Status)
					}
					n := sampleNote(5, 1)
					n.Status = sqlc.NoteStatusActive
					return n, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "note not found",
			noteID: "999",
			setup: func(q *fakeQuerier) {
				q.GetNoteByIDFn = func(_ context.Context, _ sqlc.GetNoteByIDParams) (sqlc.Note, error) {
					return sqlc.Note{}, fmt.Errorf("no rows")
				}
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid ID",
			noteID:     "bad",
			setup:      func(q *fakeQuerier) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &fakeQuerier{}
			tt.setup(q)
			h := NewNoteHandlers(q)

			r := reqWithUserAndChiCtx("POST", "/api/v1/notes/"+tt.noteID+"/promote", nil, 1, map[string]string{"id": tt.noteID})
			w := httptest.NewRecorder()

			h.Promote(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("Promote() status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var resp map[string]json.RawMessage
				decodeJSON(t, w.Body, &resp)
				if _, ok := resp["conversation_id"]; !ok {
					t.Error("response missing conversation_id")
				}
				if _, ok := resp["note"]; !ok {
					t.Error("response missing note")
				}
			}
		})
	}
}

func TestNoteHandlers_Search(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		setup      func(q *fakeQuerier)
		wantStatus int
	}{
		{
			name:  "successful search",
			query: "?q=golang",
			setup: func(q *fakeQuerier) {
				q.SearchNotesFn = func(_ context.Context, arg sqlc.SearchNotesParams) ([]sqlc.SearchNotesRow, error) {
					if arg.PlaintoTsquery != "golang" {
						t.Errorf("expected query golang, got %s", arg.PlaintoTsquery)
					}
					return []sqlc.SearchNotesRow{
						{ID: 1, Title: "Go concurrency"},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing q parameter",
			query:      "",
			setup:      func(q *fakeQuerier) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "db error returns 500",
			query: "?q=fail",
			setup: func(q *fakeQuerier) {
				q.SearchNotesFn = func(_ context.Context, _ sqlc.SearchNotesParams) ([]sqlc.SearchNotesRow, error) {
					return nil, fmt.Errorf("search error")
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &fakeQuerier{}
			tt.setup(q)
			h := NewNoteHandlers(q)

			r := reqWithUser("GET", "/api/v1/notes/search"+tt.query, nil, 1)
			w := httptest.NewRecorder()

			h.Search(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("Search() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}
