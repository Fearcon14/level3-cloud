package models

import (
	"encoding/json"
	"time"
)

// LogEntry is the response shape for GET /api/v1/instances/:id/logs.
// Each entry is either an audit log (user action) or a service log (system/async event).
type LogEntry struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`       // "audit" or "service"
	Timestamp  time.Time       `json:"timestamp"`
	Action     string          `json:"action"`     // e.g. create, update, delete, cache_get, cache_set, status_change
	Message    string          `json:"message"`   // service log message; empty for audit
	Details    json.RawMessage `json:"details"`   // audit payload (JSON object)
	Metadata   json.RawMessage `json:"metadata"`   // service log extra data (JSON object)
	TenantUser string          `json:"tenantUser,omitempty"`
	InstanceID string          `json:"instanceId"`
}
