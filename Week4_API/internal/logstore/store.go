package logstore

import (
	"context"
	"encoding/json"
	"time"
)

// Store persists audit and service logs for user-centric monitoring.
// If nil, handlers skip writing/reading logs (e.g. when DATABASE_URL is unset).
type Store interface {
	AppendAuditLog(ctx context.Context, tenantUser, instanceID, action string, details map[string]any) error
	AppendServiceLog(ctx context.Context, tenantUser, instanceID, eventType, message string, metadata map[string]any) error
	ListLogs(ctx context.Context, tenantUser, instanceID string, opts ListOpts) ([]LogEntry, error)
	// ListLogsAll returns logs for the tenant across all instances, optionally filtered by InstanceID.
	ListLogsAll(ctx context.Context, tenantUser string, opts ListOpts) ([]LogEntry, error)
}

// ListOpts filters and paginates log listing.
type ListOpts struct {
	Type       string    // "audit", "service", or "" for both
	Since      time.Time // optional; zero means no lower bound
	Limit      int       // max entries; 0 means default (e.g. 100)
	InstanceID string    // optional; when set filter by instance (for ListLogsAll)
}

// LogEntry is a single audit or service log row returned by ListLogs.
type LogEntry struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"` // "audit" or "service"
	Timestamp  time.Time       `json:"timestamp"`
	Action     string          `json:"action"`     // audit: action; service: event_type
	Message    string          `json:"message"`   // service log message; empty for audit
	Details    json.RawMessage `json:"details"`   // audit details (JSON); nil if empty
	Metadata   json.RawMessage `json:"metadata"`  // service metadata (JSON); nil if empty
	TenantUser string          `json:"tenantUser,omitempty"`
	InstanceID string          `json:"instanceId"`
}
