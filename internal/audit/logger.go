package audit

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

type auditContextKey string

const (
	actorKey auditContextKey = "audit_actor"
)

// WithActor returns a new context with the provided actor ID.
func WithActor(ctx context.Context, actor string) context.Context {
	return context.WithValue(ctx, actorKey, actor)
}

// FromContext extracts the actor ID from the context.
func FromContext(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(actorKey).(string)
	return val, ok
}

type Logger struct {
	mu       sync.Mutex
	secret   []byte
	sink     Sink
	lastHash string
}

func NewLogger(secret string, sink Sink) *Logger {
	if sink == nil {
		return nil
	}
	s := secret
	if s == "" {
		s = "default-stellabill-internal-secret" // Fallback for dev
	}
	return &Logger{
		secret: []byte(s),
		sink:   sink,
	}
}

func (l *Logger) Log(ctx context.Context, event AuditEvent) (AuditEvent, error) {
	if l == nil {
		return AuditEvent{}, errors.New("audit logger is not initialized")
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// 1. Prepare Event Metadata
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	} else {
		event.Timestamp = event.Timestamp.UTC()
	}
	
	// 2. Redaction (PII Protection)
	event.Metadata = l.redact(event.Metadata)

	// 3. Cryptographic Chaining
	event.PrevHash = l.lastHash
	event.Hash = l.computeHash(event)
	l.lastHash = event.Hash

	// 4. Persistence
	if err := l.sink.WriteEvent(event); err != nil {
		return AuditEvent{}, fmt.Errorf("failed to write to sink: %w", err)
	}

	return event, nil
}

func (l *Logger) computeHash(e AuditEvent) string {
	// Create a stable string representation for hashing
	raw := fmt.Sprintf("%d|%s|%s|%s|%s|%s|%v", 
		e.Timestamp.Unix(), e.Actor, e.Action, e.Resource, e.Outcome, e.PrevHash, e.Metadata)
	
	h := hmac.New(sha256.New, l.secret)
	h.Write([]byte(raw))
	return hex.EncodeToString(h.Sum(nil))
}

const redactedValue = "[REDACTED]"

func (l *Logger) redact(meta map[string]interface{}) map[string]interface{} {
	if meta == nil {
		return nil
	}
	
	sensitiveKeys := []string{"password", "token", "secret", "auth", "key", "cvv", "card"}
	newMeta := make(map[string]interface{})

	for k, v := range meta {
		valStr := strings.ToLower(fmt.Sprintf("%v", v))
		isSensitive := false
		
		for _, sk := range sensitiveKeys {
			if strings.Contains(strings.ToLower(k), sk) || strings.Contains(valStr, "bearer") {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			newMeta[k] = redactedValue
		} else {
			newMeta[k] = v
		}
	}
	return newMeta
}

func (l *Logger) LastHash() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.lastHash
}
<<<<<<< HEAD

func (l *Logger) computeHash(entry Entry) string {
	clone := entry
	clone.Hash = ""
	payload, _ := json.Marshal(clone)

	mac := hmac.New(sha256.New, l.secret)
	mac.Write([]byte(entry.PrevHash))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func fallbackActor(ctx context.Context, actor string) string {
	if trimmed := strings.TrimSpace(actor); trimmed != "" {
		return trimmed
	}
	if ctx != nil {
		if v := ctx.Value(actorContextKey{}); v != nil {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s)
			}
		}
	}
	return "anonymous"
}

// actorContextKey is used when an upstream authenticator wants to seed the actor identity.
type actorContextKey struct{}

// WithActor annotates a context with the actor performing the action.
func WithActor(ctx context.Context, actor string) context.Context {
	return context.WithValue(ctx, actorContextKey{}, strings.TrimSpace(actor))
}

func redact(meta map[string]string) map[string]string {
	if len(meta) == 0 {
		return nil
	}
	sanitized := make(map[string]string, len(meta))
	for k, v := range meta {
		lower := strings.ToLower(k)
		if containsSensitiveKey(lower) || looksSensitiveValue(v) {
			sanitized[k] = redactedValue
			continue
		}
		sanitized[k] = v
	}
	return sanitized
}

func containsSensitiveKey(key string) bool {
	switch {
	case strings.Contains(key, "password"),
		strings.Contains(key, "secret"),
		strings.Contains(key, "token"),
		strings.Contains(key, "authorization"),
		strings.Contains(key, "auth_header"),
		strings.Contains(key, "api_key"):
		return true
	default:
		return false
	}
}

func looksSensitiveValue(v string) bool {
	v = strings.TrimSpace(strings.ToLower(v))
	return strings.HasPrefix(v, "bearer ") || strings.HasPrefix(v, "basic ")
}

=======
>>>>>>> upstream/main
