//go:build integration

package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/aray/cerebray/backend/db/sqlc"
	"github.com/aray/cerebray/backend/internal/handlers"
	mw "github.com/aray/cerebray/backend/internal/middleware"
)

// shared container and queries across all tests in the file
var (
	testQueries *sqlc.Queries
	testPool    *pgxpool.Pool
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Find migrations directory relative to this test file
	_, thisFile, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "db", "migrations")
	absDir, err := filepath.Abs(migrationsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve migrations path: %v\n", err)
		os.Exit(1)
	}

	// Start postgres container
	pgContainer, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("cerebray_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start postgres container: %v\n", err)
		os.Exit(1)
	}
	defer pgContainer.Terminate(ctx) //nolint:errcheck

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get connection string: %v\n", err)
		os.Exit(1)
	}

	// Run migrations
	// golang-migrate pgx5 driver uses "pgx5://" scheme
	migrateConnStr := "pgx5" + connStr[len("postgres"):]
	mig, err := migrate.New("file://"+absDir, migrateConnStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create migrate instance: %v\n", err)
		os.Exit(1)
	}
	if err := mig.Up(); err != nil && err != migrate.ErrNoChange {
		fmt.Fprintf(os.Stderr, "migration failed: %v\n", err)
		os.Exit(1)
	}
	mig.Close()

	// Connect pgx pool
	testPool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect pgx pool: %v\n", err)
		os.Exit(1)
	}
	defer testPool.Close()

	testQueries = sqlc.New(testPool)

	os.Exit(m.Run())
}

// createTestUser inserts a user directly and returns the user ID.
func createTestUser(t *testing.T, suffix string) int64 {
	t.Helper()
	user, err := testQueries.CreateUser(context.Background(), sqlc.CreateUserParams{
		OidcSubject: "test-subject-" + suffix,
		Email:       "test-" + suffix + "@example.com",
		Name:        "Test User " + suffix,
	})
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return user.ID
}

// injectUserID is middleware that injects user_id into the request context,
// bypassing real auth for integration tests.
func injectUserID(userID int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), mw.UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// buildNoteRouter creates a Chi router with note handlers wired up, using
// the fake auth middleware to inject the given user ID.
func buildNoteRouter(userID int64) http.Handler {
	nh := handlers.NewNoteHandlers(testQueries)
	r := chi.NewRouter()
	r.Use(injectUserID(userID))

	r.Route("/api/v1/notes", func(r chi.Router) {
		r.Get("/", nh.List)
		r.Post("/", nh.Create)
		r.Get("/search", nh.Search)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", nh.Get)
			r.Put("/", nh.Update)
			r.Delete("/", nh.Delete)
			r.Post("/promote", nh.Promote)
			r.Post("/sleep", nh.Sleep)
			r.Post("/archive", nh.Archive)
		})
	})

	return r
}

func TestNoteCRUDLifecycle(t *testing.T) {
	userID := createTestUser(t, "crud")
	srv := httptest.NewServer(buildNoteRouter(userID))
	defer srv.Close()

	var noteID float64

	// --- Create ---
	body := map[string]string{
		"title":   "Test Concept",
		"summary": "A quick summary for testing",
		"body":    "Full body content here",
	}
	b, _ := json.Marshal(body)
	resp, err := http.Post(srv.URL+"/api/v1/notes", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST /notes failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var created map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&created)
	resp.Body.Close()

	noteID = created["id"].(float64)
	if created["title"] != "Test Concept" {
		t.Errorf("expected title 'Test Concept', got %v", created["title"])
	}
	if created["status"] != "fleeting" {
		t.Errorf("expected default status 'fleeting', got %v", created["status"])
	}

	// --- Get ---
	resp, err = http.Get(fmt.Sprintf("%s/api/v1/notes/%d", srv.URL, int64(noteID)))
	if err != nil {
		t.Fatalf("GET /notes/{id} failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var fetched map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&fetched)
	resp.Body.Close()
	if fetched["title"] != "Test Concept" {
		t.Errorf("GET returned wrong title: %v", fetched["title"])
	}

	// --- Update ---
	updateBody := map[string]string{
		"title":     "Updated Concept",
		"summary":   "Updated summary",
		"body":      "Updated body",
		"note_type": "concept",
		"status":    "fleeting",
		"tlp":       "green",
	}
	b, _ = json.Marshal(updateBody)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v1/notes/%d", srv.URL, int64(noteID)), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT /notes/{id} failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var updated map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&updated)
	resp.Body.Close()
	if updated["title"] != "Updated Concept" {
		t.Errorf("expected updated title 'Updated Concept', got %v", updated["title"])
	}
	if updated["tlp"] != "green" {
		t.Errorf("expected tlp 'green', got %v", updated["tlp"])
	}

	// --- Delete ---
	req, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v1/notes/%d", srv.URL, int64(noteID)), nil)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE /notes/{id} failed: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// --- Get after delete (expect 404) ---
	resp, err = http.Get(fmt.Sprintf("%s/api/v1/notes/%d", srv.URL, int64(noteID)))
	if err != nil {
		t.Fatalf("GET /notes/{id} after delete failed: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestNotePromotionFlow(t *testing.T) {
	userID := createTestUser(t, "promote")
	srv := httptest.NewServer(buildNoteRouter(userID))
	defer srv.Close()

	// Create a fleeting note
	body := map[string]string{
		"title":   "Fleeting Thought",
		"summary": "Something to promote",
	}
	b, _ := json.Marshal(body)
	resp, err := http.Post(srv.URL+"/api/v1/notes", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST /notes failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	var created map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&created)
	resp.Body.Close()

	noteID := int64(created["id"].(float64))
	if created["status"] != "fleeting" {
		t.Fatalf("expected initial status 'fleeting', got %v", created["status"])
	}

	// Promote the note
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v1/notes/%d/promote", srv.URL, noteID), nil)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /promote failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var promoteResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&promoteResult)
	resp.Body.Close()

	// The promote response includes a nested "note" and "conversation_id"
	noteData, ok := promoteResult["note"].(map[string]interface{})
	if !ok {
		t.Fatalf("promote response missing 'note' field")
	}
	if noteData["status"] != "active" {
		t.Errorf("expected promoted status 'active', got %v", noteData["status"])
	}

	convoID, ok := promoteResult["conversation_id"]
	if !ok || convoID == nil {
		t.Fatal("promote response missing 'conversation_id'")
	}
	if convoID.(float64) <= 0 {
		t.Errorf("expected positive conversation_id, got %v", convoID)
	}

	// Verify the note now has source_chat_id set
	resp, err = http.Get(fmt.Sprintf("%s/api/v1/notes/%d", srv.URL, noteID))
	if err != nil {
		t.Fatalf("GET after promote failed: %v", err)
	}
	var afterPromote map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&afterPromote)
	resp.Body.Close()
	if afterPromote["source_chat_id"] == nil {
		t.Error("expected source_chat_id to be set after promotion")
	}
}

func TestNoteSearch(t *testing.T) {
	userID := createTestUser(t, "search")
	srv := httptest.NewServer(buildNoteRouter(userID))
	defer srv.Close()

	// Create 3 notes with different content
	notes := []map[string]string{
		{"title": "Quantum Entanglement", "summary": "Particles linked across distance", "body": "Spooky action at a distance"},
		{"title": "Neural Networks", "summary": "Machine learning architecture", "body": "Layers of interconnected nodes"},
		{"title": "Quantum Computing", "summary": "Computing with qubits", "body": "Superposition and entanglement based computation"},
	}

	for _, n := range notes {
		b, _ := json.Marshal(n)
		resp, err := http.Post(srv.URL+"/api/v1/notes", "application/json", bytes.NewReader(b))
		if err != nil {
			t.Fatalf("POST /notes failed: %v", err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
		resp.Body.Close()
	}

	// Search for "quantum" - should match 2 notes
	resp, err := http.Get(srv.URL + "/api/v1/notes/search?q=quantum")
	if err != nil {
		t.Fatalf("GET /search failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var results []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&results)
	resp.Body.Close()

	if len(results) != 2 {
		t.Errorf("expected 2 search results for 'quantum', got %d", len(results))
	}

	// Search for "neural" - should match 1 note
	resp, err = http.Get(srv.URL + "/api/v1/notes/search?q=neural")
	if err != nil {
		t.Fatalf("GET /search failed: %v", err)
	}
	var neuralResults []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&neuralResults)
	resp.Body.Close()

	if len(neuralResults) != 1 {
		t.Errorf("expected 1 search result for 'neural', got %d", len(neuralResults))
	}

	// Search with no query param - should return 400
	resp, err = http.Get(srv.URL + "/api/v1/notes/search")
	if err != nil {
		t.Fatalf("GET /search without q failed: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing q param, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestNoteStatusTransitions(t *testing.T) {
	userID := createTestUser(t, "transitions")
	srv := httptest.NewServer(buildNoteRouter(userID))
	defer srv.Close()

	// Create a note
	b, _ := json.Marshal(map[string]string{"title": "Status Test Note"})
	resp, err := http.Post(srv.URL+"/api/v1/notes", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST /notes failed: %v", err)
	}
	var created map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&created)
	resp.Body.Close()
	noteID := int64(created["id"].(float64))

	// Sleep the note
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v1/notes/%d/sleep", srv.URL, noteID), nil)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /sleep failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from /sleep, got %d", resp.StatusCode)
	}
	var sleepResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&sleepResult)
	resp.Body.Close()
	if sleepResult["status"] != "sleeping" {
		t.Errorf("expected status 'sleeping', got %v", sleepResult["status"])
	}

	// Archive the note
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v1/notes/%d/archive", srv.URL, noteID), nil)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /archive failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from /archive, got %d", resp.StatusCode)
	}
	var archiveResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&archiveResult)
	resp.Body.Close()
	if archiveResult["status"] != "archived" {
		t.Errorf("expected status 'archived', got %v", archiveResult["status"])
	}
}

func TestNoteListFiltering(t *testing.T) {
	userID := createTestUser(t, "listing")
	srv := httptest.NewServer(buildNoteRouter(userID))
	defer srv.Close()

	// Create notes with different statuses
	for _, title := range []string{"Note A", "Note B", "Note C"} {
		b, _ := json.Marshal(map[string]string{"title": title})
		resp, _ := http.Post(srv.URL+"/api/v1/notes", "application/json", bytes.NewReader(b))
		resp.Body.Close()
	}

	// List all notes
	resp, err := http.Get(srv.URL + "/api/v1/notes")
	if err != nil {
		t.Fatalf("GET /notes failed: %v", err)
	}
	var allNotes []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&allNotes)
	resp.Body.Close()
	if len(allNotes) < 3 {
		t.Errorf("expected at least 3 notes, got %d", len(allNotes))
	}

	// List by status=fleeting
	resp, err = http.Get(srv.URL + "/api/v1/notes?status=fleeting")
	if err != nil {
		t.Fatalf("GET /notes?status=fleeting failed: %v", err)
	}
	var fleetingNotes []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&fleetingNotes)
	resp.Body.Close()
	for _, n := range fleetingNotes {
		if n["status"] != "fleeting" {
			t.Errorf("expected all results to have status 'fleeting', got %v", n["status"])
		}
	}
}

func TestCreateNoteValidation(t *testing.T) {
	userID := createTestUser(t, "validation")
	srv := httptest.NewServer(buildNoteRouter(userID))
	defer srv.Close()

	// Missing title - should return 400
	b, _ := json.Marshal(map[string]string{"summary": "no title here"})
	resp, err := http.Post(srv.URL+"/api/v1/notes", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST /notes failed: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing title, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Invalid JSON - should return 400
	resp, err = http.Post(srv.URL+"/api/v1/notes", "application/json", bytes.NewReader([]byte("not json")))
	if err != nil {
		t.Fatalf("POST /notes failed: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}
