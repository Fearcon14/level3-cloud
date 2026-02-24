package logstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const defaultListLimit = 100

// PostgresStore implements Store using PostgreSQL (audit_logs and service_logs tables).
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore opens a connection and returns a PostgresStore. Caller must call Close when done.
func NewPostgresStore(ctx context.Context, databaseURL string) (*PostgresStore, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return &PostgresStore{db: db}, nil
}

// Close closes the database connection.
func (s *PostgresStore) Close() error {
	return s.db.Close()
}

// AppendAuditLog inserts one audit log row.
func (s *PostgresStore) AppendAuditLog(ctx context.Context, tenantUser, instanceID, action string, details map[string]any) error {
	var detailsJSON []byte
	if details != nil {
		var err error
		detailsJSON, err = json.Marshal(details)
		if err != nil {
			return err
		}
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO audit_logs (tenant_user, instance_id, action, details) VALUES ($1, $2, $3, $4)`,
		tenantUser, instanceID, action, detailsJSON)
	return err
}

// AppendServiceLog inserts one service log row.
func (s *PostgresStore) AppendServiceLog(ctx context.Context, tenantUser, instanceID, eventType, message string, metadata map[string]any) error {
	var metadataJSON []byte
	if metadata != nil {
		var err error
		metadataJSON, err = json.Marshal(metadata)
		if err != nil {
			return err
		}
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO service_logs (tenant_user, instance_id, event_type, message, metadata) VALUES ($1, $2, $3, $4, $5)`,
		tenantUser, instanceID, eventType, message, metadataJSON)
	return err
}

// ListLogs returns audit and/or service logs for the instance, tenant-scoped, ordered by time desc.
func (s *PostgresStore) ListLogs(ctx context.Context, tenantUser, instanceID string, opts ListOpts) ([]LogEntry, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}

	var entries []LogEntry

	switch opts.Type {
	case "audit":
		rows, err := s.queryAuditLogs(ctx, tenantUser, instanceID, opts.Since, limit)
		if err != nil {
			return nil, err
		}
		entries = rows
	case "service":
		rows, err := s.queryServiceLogs(ctx, tenantUser, instanceID, opts.Since, limit)
		if err != nil {
			return nil, err
		}
		entries = rows
	default:
		// both: fetch audit and service, merge by created_at desc, then take limit
		auditRows, err := s.queryAuditLogs(ctx, tenantUser, instanceID, opts.Since, limit*2)
		if err != nil {
			return nil, err
		}
		serviceRows, err := s.queryServiceLogs(ctx, tenantUser, instanceID, opts.Since, limit*2)
		if err != nil {
			return nil, err
		}
		entries = mergeLogEntries(auditRows, serviceRows, limit)
	}

	return entries, nil
}

func (s *PostgresStore) queryAuditLogs(ctx context.Context, tenantUser, instanceID string, since time.Time, limit int) ([]LogEntry, error) {
	var rows *sql.Rows
	var err error
	if since.IsZero() {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, created_at, action, COALESCE(details::text, 'null') FROM audit_logs WHERE tenant_user = $1 AND instance_id = $2 ORDER BY created_at DESC LIMIT $3`,
			tenantUser, instanceID, limit)
	} else {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, created_at, action, COALESCE(details::text, 'null') FROM audit_logs WHERE tenant_user = $1 AND instance_id = $2 AND created_at >= $3 ORDER BY created_at DESC LIMIT $4`,
			tenantUser, instanceID, since, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []LogEntry
	for rows.Next() {
		var id string
		var createdAt time.Time
		var action string
		var detailsStr string
		if err := rows.Scan(&id, &createdAt, &action, &detailsStr); err != nil {
			return nil, err
		}
		e := LogEntry{
			ID:         id,
			Type:       "audit",
			Timestamp:  createdAt,
			Action:     action,
			TenantUser: tenantUser,
			InstanceID: instanceID,
			Details:    json.RawMessage(detailsStr),
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *PostgresStore) queryServiceLogs(ctx context.Context, tenantUser, instanceID string, since time.Time, limit int) ([]LogEntry, error) {
	var rows *sql.Rows
	var err error
	if since.IsZero() {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, created_at, event_type, message, COALESCE(metadata::text, 'null') FROM service_logs WHERE tenant_user = $1 AND instance_id = $2 ORDER BY created_at DESC LIMIT $3`,
			tenantUser, instanceID, limit)
	} else {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, created_at, event_type, message, COALESCE(metadata::text, 'null') FROM service_logs WHERE tenant_user = $1 AND instance_id = $2 AND created_at >= $3 ORDER BY created_at DESC LIMIT $4`,
			tenantUser, instanceID, since, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []LogEntry
	for rows.Next() {
		var id string
		var createdAt time.Time
		var eventType string
		var message sql.NullString
		var metadataStr string
		if err := rows.Scan(&id, &createdAt, &eventType, &message, &metadataStr); err != nil {
			return nil, err
		}
		e := LogEntry{
			ID:         id,
			Type:       "service",
			Timestamp:  createdAt,
			Action:     eventType,
			Message:    message.String,
			TenantUser: tenantUser,
			InstanceID: instanceID,
			Metadata:   json.RawMessage(metadataStr),
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// ListLogsAll returns audit and/or service logs for the tenant, optionally filtered by instance, ordered by time desc.
func (s *PostgresStore) ListLogsAll(ctx context.Context, tenantUser string, opts ListOpts) ([]LogEntry, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}

	var entries []LogEntry

	switch opts.Type {
	case "audit":
		rows, err := s.queryAuditLogsAll(ctx, tenantUser, opts.Since, opts.InstanceID, limit)
		if err != nil {
			return nil, err
		}
		entries = rows
	case "service":
		rows, err := s.queryServiceLogsAll(ctx, tenantUser, opts.Since, opts.InstanceID, limit)
		if err != nil {
			return nil, err
		}
		entries = rows
	default:
		auditRows, err := s.queryAuditLogsAll(ctx, tenantUser, opts.Since, opts.InstanceID, limit*2)
		if err != nil {
			return nil, err
		}
		serviceRows, err := s.queryServiceLogsAll(ctx, tenantUser, opts.Since, opts.InstanceID, limit*2)
		if err != nil {
			return nil, err
		}
		entries = mergeLogEntries(auditRows, serviceRows, limit)
	}

	return entries, nil
}

func (s *PostgresStore) queryAuditLogsAll(ctx context.Context, tenantUser string, since time.Time, instanceID string, limit int) ([]LogEntry, error) {
	var rows *sql.Rows
	var err error
	if instanceID != "" {
		if since.IsZero() {
			rows, err = s.db.QueryContext(ctx,
				`SELECT id, instance_id, created_at, action, COALESCE(details::text, 'null') FROM audit_logs WHERE tenant_user = $1 AND instance_id = $2 ORDER BY created_at DESC LIMIT $3`,
				tenantUser, instanceID, limit)
		} else {
			rows, err = s.db.QueryContext(ctx,
				`SELECT id, instance_id, created_at, action, COALESCE(details::text, 'null') FROM audit_logs WHERE tenant_user = $1 AND instance_id = $2 AND created_at >= $3 ORDER BY created_at DESC LIMIT $4`,
				tenantUser, instanceID, since, limit)
		}
	} else {
		if since.IsZero() {
			rows, err = s.db.QueryContext(ctx,
				`SELECT id, instance_id, created_at, action, COALESCE(details::text, 'null') FROM audit_logs WHERE tenant_user = $1 ORDER BY created_at DESC LIMIT $2`,
				tenantUser, limit)
		} else {
			rows, err = s.db.QueryContext(ctx,
				`SELECT id, instance_id, created_at, action, COALESCE(details::text, 'null') FROM audit_logs WHERE tenant_user = $1 AND created_at >= $2 ORDER BY created_at DESC LIMIT $3`,
				tenantUser, since, limit)
		}
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []LogEntry
	for rows.Next() {
		var id, instID string
		var createdAt time.Time
		var action string
		var detailsStr string
		if err := rows.Scan(&id, &instID, &createdAt, &action, &detailsStr); err != nil {
			return nil, err
		}
		e := LogEntry{
			ID:         id,
			Type:       "audit",
			Timestamp:  createdAt,
			Action:     action,
			TenantUser: tenantUser,
			InstanceID: instID,
			Details:    json.RawMessage(detailsStr),
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *PostgresStore) queryServiceLogsAll(ctx context.Context, tenantUser string, since time.Time, instanceID string, limit int) ([]LogEntry, error) {
	var rows *sql.Rows
	var err error
	if instanceID != "" {
		if since.IsZero() {
			rows, err = s.db.QueryContext(ctx,
				`SELECT id, instance_id, created_at, event_type, message, COALESCE(metadata::text, 'null') FROM service_logs WHERE tenant_user = $1 AND instance_id = $2 ORDER BY created_at DESC LIMIT $3`,
				tenantUser, instanceID, limit)
		} else {
			rows, err = s.db.QueryContext(ctx,
				`SELECT id, instance_id, created_at, event_type, message, COALESCE(metadata::text, 'null') FROM service_logs WHERE tenant_user = $1 AND instance_id = $2 AND created_at >= $3 ORDER BY created_at DESC LIMIT $4`,
				tenantUser, instanceID, since, limit)
		}
	} else {
		if since.IsZero() {
			rows, err = s.db.QueryContext(ctx,
				`SELECT id, instance_id, created_at, event_type, message, COALESCE(metadata::text, 'null') FROM service_logs WHERE tenant_user = $1 ORDER BY created_at DESC LIMIT $2`,
				tenantUser, limit)
		} else {
			rows, err = s.db.QueryContext(ctx,
				`SELECT id, instance_id, created_at, event_type, message, COALESCE(metadata::text, 'null') FROM service_logs WHERE tenant_user = $1 AND created_at >= $2 ORDER BY created_at DESC LIMIT $3`,
				tenantUser, since, limit)
		}
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []LogEntry
	for rows.Next() {
		var id, instID string
		var createdAt time.Time
		var eventType string
		var message sql.NullString
		var metadataStr string
		if err := rows.Scan(&id, &instID, &createdAt, &eventType, &message, &metadataStr); err != nil {
			return nil, err
		}
		e := LogEntry{
			ID:         id,
			Type:       "service",
			Timestamp:  createdAt,
			Action:     eventType,
			Message:    message.String,
			TenantUser: tenantUser,
			InstanceID: instID,
			Metadata:   json.RawMessage(metadataStr),
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// mergeLogEntries interleaves audit and service entries by timestamp desc and returns up to limit.
func mergeLogEntries(audit, service []LogEntry, limit int) []LogEntry {
	i, j := 0, 0
	var out []LogEntry
	for len(out) < limit && (i < len(audit) || j < len(service)) {
		if i >= len(audit) {
			out = append(out, service[j])
			j++
			continue
		}
		if j >= len(service) {
			out = append(out, audit[i])
			i++
			continue
		}
		if audit[i].Timestamp.After(service[j].Timestamp) {
			out = append(out, audit[i])
			i++
		} else {
			out = append(out, service[j])
			j++
		}
	}
	return out
}
