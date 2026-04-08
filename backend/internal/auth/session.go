package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrSessionNotFound   = errors.New("session not found")
	ErrInvalidStateToken = errors.New("invalid or expired state token")
)

type sessionData struct {
	UserID int64 `json:"user_id"`
}

type stateData struct {
	State    string `json:"state"`
	Verifier string `json:"verifier"`
}

// SessionStore manages user sessions and OAuth state tokens in Redis.
type SessionStore struct {
	client     *redis.Client
	sessionTTL time.Duration
	stateTTL   time.Duration
}

func NewSessionStore(client *redis.Client) *SessionStore {
	return &SessionStore{
		client:     client,
		sessionTTL: SessionTTL,
		stateTTL:   StateTTL,
	}
}

func (s *SessionStore) Create(ctx context.Context, userID int64) (string, error) {
	sessionID, err := GenerateRandomToken(32)
	if err != nil {
		return "", fmt.Errorf("generating session ID: %w", err)
	}

	data, err := json.Marshal(sessionData{UserID: userID})
	if err != nil {
		return "", fmt.Errorf("marshalling session: %w", err)
	}

	key := "session:" + sessionID
	if err := s.client.Set(ctx, key, data, s.sessionTTL).Err(); err != nil {
		return "", fmt.Errorf("storing session: %w", err)
	}

	return sessionID, nil
}

func (s *SessionStore) Get(ctx context.Context, sessionID string) (int64, error) {
	key := "session:" + sessionID
	val, err := s.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return 0, ErrSessionNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("reading session: %w", err)
	}

	var sd sessionData
	if err := json.Unmarshal([]byte(val), &sd); err != nil {
		return 0, fmt.Errorf("unmarshalling session: %w", err)
	}

	// Refresh TTL on access (sliding expiry)
	s.client.Expire(ctx, key, s.sessionTTL)

	return sd.UserID, nil
}

func (s *SessionStore) Delete(ctx context.Context, sessionID string) error {
	return s.client.Del(ctx, "session:"+sessionID).Err()
}

func (s *SessionStore) StoreStateToken(ctx context.Context, state, verifier string) error {
	data, err := json.Marshal(stateData{State: state, Verifier: verifier})
	if err != nil {
		return fmt.Errorf("marshalling state: %w", err)
	}
	key := "oauth_state:" + state
	return s.client.Set(ctx, key, data, s.stateTTL).Err()
}

func (s *SessionStore) ValidateAndConsumeStateToken(ctx context.Context, state string) (string, error) {
	key := "oauth_state:" + state
	val, err := s.client.GetDel(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrInvalidStateToken
	}
	if err != nil {
		return "", fmt.Errorf("reading state token: %w", err)
	}

	var sd stateData
	if err := json.Unmarshal([]byte(val), &sd); err != nil {
		return "", fmt.Errorf("unmarshalling state: %w", err)
	}

	return sd.Verifier, nil
}
