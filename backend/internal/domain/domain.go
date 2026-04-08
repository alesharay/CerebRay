package domain

import (
	"encoding/json"
	"time"
)

// User represents an authenticated user.
type User struct {
	ID          int64            `json:"id"`
	OIDCSubject string           `json:"oidc_subject"`
	Email       string           `json:"email"`
	Name        string           `json:"name"`
	AvatarURL   string           `json:"avatar_url"`
	Preferences json.RawMessage  `json:"preferences"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// NoteType defines the purpose of a note.
type NoteType string

const (
	NoteTypeConcept   NoteType = "concept"
	NoteTypeTheory    NoteType = "theory"
	NoteTypeInsight   NoteType = "insight"
	NoteTypeQuote     NoteType = "quote"
	NoteTypeReference NoteType = "reference"
	NoteTypeQuestion  NoteType = "question"
	NoteTypeStructure NoteType = "structure"
	NoteTypeGuide     NoteType = "guide"
)

// NoteStatus describes where a note is in the processing pipeline.
type NoteStatus string

const (
	NoteStatusFleeting NoteStatus = "fleeting"
	NoteStatusSleeping NoteStatus = "sleeping"
	NoteStatusActive   NoteStatus = "active"
	NoteStatusLinked   NoteStatus = "linked"
	NoteStatusArchived NoteStatus = "archived"
)

// NoteTLP is the Traffic Light Protocol level.
type NoteTLP string

const (
	NoteTLPClear NoteTLP = "clear"
	NoteTLPGreen NoteTLP = "green"
	NoteTLPAmber NoteTLP = "amber"
	NoteTLPRed   NoteTLP = "red"
)

// Note represents a Zettelkasten note.
type Note struct {
	ID            int64      `json:"id"`
	UserID        int64      `json:"user_id"`
	Title         string     `json:"title"`
	Summary       string     `json:"summary"`
	LaymansTerms  string     `json:"laymans_terms"`
	Analogy       string     `json:"analogy"`
	CoreIdea      string     `json:"core_idea"`
	Body          string     `json:"body"`
	Components    string     `json:"components"`
	WhyItMatters  string     `json:"why_it_matters"`
	Examples      string     `json:"examples"`
	Templates     string     `json:"templates"`
	Additional    string     `json:"additional"`
	NoteType      NoteType   `json:"note_type"`
	Status        NoteStatus `json:"status"`
	TLP           NoteTLP    `json:"tlp"`
	SourceChatID  *int64     `json:"source_chat_id,omitempty"`
	Tags          []string   `json:"tags,omitempty"`
	ConnectionCount int      `json:"connection_count"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// Connection links two notes together.
type Connection struct {
	ID             int64     `json:"id"`
	SourceID       int64     `json:"source_id"`
	TargetID       int64     `json:"target_id"`
	Label          string    `json:"label"`
	ConnectedID    int64     `json:"connected_id"`
	ConnectedTitle string    `json:"connected_title"`
	CreatedAt      time.Time `json:"created_at"`
}

// Tag is a keyword associated with notes.
type Tag struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	Name      string `json:"name"`
	NoteCount int64  `json:"note_count"`
}

// Conversation is an AI chat session.
type Conversation struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Title     string    `json:"title"`
	Topic     string    `json:"topic"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Message is a single message within a conversation.
type Message struct {
	ID             int64     `json:"id"`
	ConversationID int64     `json:"conversation_id"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	InputTokens    int       `json:"input_tokens"`
	OutputTokens   int       `json:"output_tokens"`
	Model          string    `json:"model"`
	CreatedAt      time.Time `json:"created_at"`
}

// GlossaryTerm is a definition in the glossary.
type GlossaryTerm struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	Term            string    `json:"term"`
	Definition      string    `json:"definition"`
	SourceNoteID    *int64    `json:"source_note_id,omitempty"`
	SourceNoteTitle string    `json:"source_note_title,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// DashboardStats holds zone counts for the dashboard.
type DashboardStats struct {
	InboxCount  int64 `json:"inbox_count"`
	EchoesCount int64 `json:"echoes_count"`
	CodexCount  int64 `json:"codex_count"`
}

// GraphData holds nodes and edges for the Index/MOC visualization.
type GraphNode struct {
	ID    int64    `json:"id"`
	Title string   `json:"title"`
	Type  NoteType `json:"type"`
}

type GraphEdge struct {
	Source int64  `json:"source"`
	Target int64  `json:"target"`
	Label  string `json:"label"`
}

type GraphData struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}
