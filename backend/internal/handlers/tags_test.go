package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/aray/cerebray/backend/db/sqlc"
)

// The frontend posts {"name":"..."} to /notes/{id}/tags. The handler used to
// decode {"tags":["..."]}, so the slice came back nil, the create loop never
// ran, and it still returned 204. Tagging looked fine and silently wrote
// nothing for months. These tests pin the wire shape and make the failure loud.
func TestTagHandlers_AddToNote(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantCreate bool
		wantName   string
	}{
		{
			name:       "creates tag from name field",
			body:       `{"name":"zettelkasten"}`,
			wantStatus: http.StatusNoContent,
			wantCreate: true,
			wantName:   "zettelkasten",
		},
		{
			name:       "trims surrounding whitespace",
			body:       `{"name":"  spaced  "}`,
			wantStatus: http.StatusNoContent,
			wantCreate: true,
			wantName:   "spaced",
		},
		{
			name:       "empty name is rejected, not silently ignored",
			body:       `{"name":""}`,
			wantStatus: http.StatusBadRequest,
			wantCreate: false,
		},
		{
			name:       "missing name is rejected",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
			wantCreate: false,
		},
		{
			// The old wire shape must not quietly succeed any more.
			name:       "legacy tags array is rejected",
			body:       `{"tags":["old","shape"]}`,
			wantStatus: http.StatusBadRequest,
			wantCreate: false,
		},
		{
			name:       "malformed json is rejected",
			body:       `{`,
			wantStatus: http.StatusBadRequest,
			wantCreate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var created bool
			var gotName string
			var linked bool

			q := &fakeQuerier{}
			q.GetNoteByIDFn = func(_ context.Context, arg sqlc.GetNoteByIDParams) (sqlc.GetNoteByIDRow, error) {
				if arg.UserID != 1 {
					t.Errorf("ownership check userID = %d, want 1", arg.UserID)
				}
				return sampleNoteByIDRow(arg.ID, arg.UserID), nil
			}
			q.CreateTagFn = func(_ context.Context, arg sqlc.CreateTagParams) (sqlc.Tag, error) {
				created = true
				gotName = arg.Name
				if arg.UserID != 1 {
					t.Errorf("CreateTag userID = %d, want 1", arg.UserID)
				}
				return sqlc.Tag{ID: 7, UserID: arg.UserID, Name: arg.Name}, nil
			}
			q.AddNoteTagFn = func(_ context.Context, arg sqlc.AddNoteTagParams) (int64, error) {
				linked = true
				if arg.NoteID != 42 {
					t.Errorf("AddNoteTag noteID = %d, want 42", arg.NoteID)
				}
				if arg.TagID != 7 {
					t.Errorf("AddNoteTag tagID = %d, want 7", arg.TagID)
				}
				if arg.UserID != 1 {
					t.Errorf("AddNoteTag userID = %d, want 1 (unscoped insert)", arg.UserID)
				}
				return arg.NoteID, nil
			}

			h := NewTagHandlers(q)
			r := reqWithUserAndChiCtx("POST", "/api/v1/notes/42/tags",
				strings.NewReader(tt.body), 1, map[string]string{"id": "42"})
			w := httptest.NewRecorder()

			h.AddToNote(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("AddToNote() status = %d, want %d", w.Code, tt.wantStatus)
			}
			if created != tt.wantCreate {
				t.Errorf("CreateTag called = %v, want %v", created, tt.wantCreate)
			}
			if tt.wantCreate {
				if gotName != tt.wantName {
					t.Errorf("CreateTag name = %q, want %q", gotName, tt.wantName)
				}
				if !linked {
					t.Error("AddNoteTag was not called, tag would not attach to the note")
				}
			}
		})
	}
}

// AddNoteTag originally took only {note_id, tag_id}, so any authenticated user
// could attach a tag to anyone's note. Ownership is now checked before any
// write happens, and the insert carries user_id as a second line of defence.
func TestTagHandlers_AddToNote_RejectsForeignNote(t *testing.T) {
	var created, linked bool

	q := &fakeQuerier{}
	q.GetNoteByIDFn = func(_ context.Context, _ sqlc.GetNoteByIDParams) (sqlc.GetNoteByIDRow, error) {
		// Note 42 belongs to someone else, so the scoped lookup finds nothing.
		return sqlc.GetNoteByIDRow{}, pgx.ErrNoRows
	}
	q.CreateTagFn = func(_ context.Context, arg sqlc.CreateTagParams) (sqlc.Tag, error) {
		created = true
		return sqlc.Tag{ID: 7, UserID: arg.UserID, Name: arg.Name}, nil
	}
	q.AddNoteTagFn = func(_ context.Context, arg sqlc.AddNoteTagParams) (int64, error) {
		linked = true
		return arg.NoteID, nil
	}

	h := NewTagHandlers(q)
	r := reqWithUserAndChiCtx("POST", "/api/v1/notes/42/tags",
		strings.NewReader(`{"name":"stolen"}`), 99, map[string]string{"id": "42"})
	w := httptest.NewRecorder()

	h.AddToNote(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("AddToNote() on a foreign note status = %d, want %d", w.Code, http.StatusNotFound)
	}
	if created {
		t.Error("CreateTag was called for a note the caller does not own")
	}
	if linked {
		t.Error("AddNoteTag was called for a note the caller does not own")
	}
}

// Tag removal is scoped by user_id in SQL, so the handler has to pass the
// caller through. Without it the param defaults to 0 and the delete silently
// matches nothing.
func TestTagHandlers_RemoveFromNote_ScopesToUser(t *testing.T) {
	var got sqlc.RemoveNoteTagParams

	q := &fakeQuerier{}
	q.RemoveNoteTagFn = func(_ context.Context, arg sqlc.RemoveNoteTagParams) error {
		got = arg
		return nil
	}

	h := NewTagHandlers(q)
	r := reqWithUserAndChiCtx("DELETE", "/api/v1/notes/42/tags/7", nil, 99,
		map[string]string{"id": "42", "tagId": "7"})
	w := httptest.NewRecorder()

	h.RemoveFromNote(w, r)

	if w.Code != http.StatusNoContent {
		t.Fatalf("RemoveFromNote() status = %d, want %d", w.Code, http.StatusNoContent)
	}
	if got.UserID != 99 {
		t.Errorf("RemoveNoteTag userID = %d, want 99 (unscoped delete)", got.UserID)
	}
	if got.NoteID != 42 || got.TagID != 7 {
		t.Errorf("RemoveNoteTag noteID/tagID = %d/%d, want 42/7", got.NoteID, got.TagID)
	}
}
