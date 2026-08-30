package webhooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// EventType represents the type of event
type EventType string

const (
	EventDecisionCreated  EventType = "decision.created"
	EventDiscoveryCreated EventType = "discovery.created"
	EventFailureCreated   EventType = "failure.created"
	EventConstraintCreated EventType = "constraint.created"
	EventSessionCreated   EventType = "session.created"
	EventSessionClosed    EventType = "session.closed"
	EventCheckpointCreated EventType = "checkpoint.created"
)

// Webhook represents a webhook subscription
type Webhook struct {
	ID         string      `json:"id"`
	URL        string      `json:"url"`
	Events     []EventType `json:"events"`
	Secret     string      `json:"secret"`
	Active     bool        `json:"active"`
	CreatedAt  time.Time   `json:"created_at"`
	LastFired  *time.Time  `json:"last_fired,omitempty"`
	FailCount  int         `json:"fail_count"`
	MaxRetries int         `json:"max_retries"`
}

// Event represents an event to be sent to webhooks
type Event struct {
	ID        string      `json:"id"`
	Type      EventType   `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// WebhookPayload is what gets sent to the webhook
type WebhookPayload struct {
	Event     Event  `json:"event"`
	Timestamp int64  `json:"timestamp"`
	Signature string `json:"signature"`
}

// Manager manages webhooks
type Manager struct {
	mu            sync.RWMutex
	webhooks      map[string]*Webhook
	queue         chan Event
	httpClient    *http.Client
	maxRetries    int
	retryInterval time.Duration
}

// NewManager creates a new webhook manager
func NewManager(queueSize int, maxRetries int) *Manager {
	m := &Manager{
		webhooks:      make(map[string]*Webhook),
		queue:         make(chan Event, queueSize),
		httpClient:    &http.Client{Timeout: 10 * time.Second},
		maxRetries:    maxRetries,
		retryInterval: 5 * time.Second,
	}

	// Start event processor
	go m.processEvents()

	return m
}

// Register registers a new webhook
func (m *Manager) Register(webhook *Webhook) error {
	if webhook.ID == "" || webhook.URL == "" {
		return fmt.Errorf("webhook ID and URL are required")
	}

	if len(webhook.Events) == 0 {
		return fmt.Errorf("webhook must subscribe to at least one event")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	webhook.CreatedAt = time.Now()
	webhook.Active = true
	if webhook.MaxRetries == 0 {
		webhook.MaxRetries = m.maxRetries
	}

	m.webhooks[webhook.ID] = webhook
	return nil
}

// Unregister removes a webhook
func (m *Manager) Unregister(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.webhooks[id]; !exists {
		return fmt.Errorf("webhook not found: %s", id)
	}

	delete(m.webhooks, id)
	return nil
}

// Trigger fires an event to matching webhooks
func (m *Manager) Trigger(event Event) {
	select {
	case m.queue <- event:
	default:
		// Queue is full, drop event
		fmt.Printf("Warning: webhook queue is full, dropping event: %s\n", event.Type)
	}
}

// processEvents handles event delivery to webhooks
func (m *Manager) processEvents() {
	for event := range m.queue {
		m.mu.RLock()
		webhooks := make([]*Webhook, 0)
		for _, wh := range m.webhooks {
			if wh.Active && m.shouldTrigger(wh, event.Type) {
				webhooks = append(webhooks, wh)
			}
		}
		m.mu.RUnlock()

		// Send to matching webhooks
		for _, wh := range webhooks {
			go m.deliverEvent(wh, event)
		}
	}
}

// shouldTrigger checks if a webhook should be triggered for an event
func (m *Manager) shouldTrigger(wh *Webhook, eventType EventType) bool {
	for _, et := range wh.Events {
		if et == eventType {
			return true
		}
	}
	return false
}

// deliverEvent sends an event to a webhook with retries
func (m *Manager) deliverEvent(wh *Webhook, event Event) {
	payload := WebhookPayload{
		Event:     event,
		Timestamp: time.Now().Unix(),
	}

	// Generate signature
	payload.Signature = m.generateSignature(payload, wh.Secret)

	data, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshaling webhook payload: %v\n", err)
		return
	}

	var lastErr error
	for attempt := 0; attempt <= wh.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * m.retryInterval)
		}

		if m.sendWebhook(wh.URL, data) {
			// Success
			m.mu.Lock()
			wh.LastFired = timePtr(time.Now())
			wh.FailCount = 0
			m.mu.Unlock()
			return
		}

		lastErr = fmt.Errorf("webhook delivery failed")
	}

	// Mark as failed
	m.mu.Lock()
	wh.FailCount++
	if wh.FailCount > wh.MaxRetries {
		wh.Active = false // Disable webhook after too many failures
	}
	m.mu.Unlock()

	fmt.Printf("Failed to deliver webhook %s after %d attempts: %v\n", wh.ID, wh.MaxRetries+1, lastErr)
}

// sendWebhook sends a single webhook request
func (m *Manager) sendWebhook(url string, data []byte) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return false
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Agent-Ledger-Webhooks/1.0")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Accept 2xx responses
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// generateSignature creates an HMAC signature for webhook verification
func (m *Manager) generateSignature(payload WebhookPayload, secret string) string {
	data, _ := json.Marshal(payload.Event)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// VerifySignature verifies a webhook signature
func VerifySignature(payload []byte, signature, secret string) bool {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	expected := hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expected))
}

// List returns all registered webhooks
func (m *Manager) List() []*Webhook {
	m.mu.RLock()
	defer m.mu.RUnlock()

	webhooks := make([]*Webhook, 0, len(m.webhooks))
	for _, wh := range m.webhooks {
		webhooks = append(webhooks, wh)
	}
	return webhooks
}

// Get returns a specific webhook
func (m *Manager) Get(id string) (*Webhook, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	wh, exists := m.webhooks[id]
	if !exists {
		return nil, fmt.Errorf("webhook not found: %s", id)
	}
	return wh, nil
}

// Close gracefully shuts down the webhook manager
func (m *Manager) Close() {
	close(m.queue)
	// Wait for queue to be processed
	time.Sleep(1 * time.Second)
}

// Helper function to create time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}
