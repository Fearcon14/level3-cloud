package logstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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

// CountLogs returns the total number of log entries matching the filters (no limit/offset).
func (s *PostgresStore) CountLogs(ctx context.Context, tenantUser string, opts ListOpts) (int, error) {
	switch opts.Type {
	case "audit":
		return s.countAuditLogsAll(ctx, tenantUser, opts.Since, opts.InstanceID)
	case "service":
		return s.countServiceLogsAll(ctx, tenantUser, opts.Since, opts.InstanceID)
	default:
		auditCount, err := s.countAuditLogsAll(ctx, tenantUser, opts.Since, opts.InstanceID)
		if err != nil {
			return 0, err
		}
		serviceCount, err := s.countServiceLogsAll(ctx, tenantUser, opts.Since, opts.InstanceID)
		if err != nil {
			return 0, err
		}
		return auditCount + serviceCount, nil
	}
}

func (s *PostgresStore) countAuditLogsAll(ctx context.Context, tenantUser string, since time.Time, instanceID string) (int, error) {
	var count int
	var err error
	if instanceID != "" {
		if since.IsZero() {
			err = s.db.QueryRowContext(ctx,
				`SELECT COUNT(*) FROM audit_logs WHERE tenant_user = $1 AND instance_id = $2`,
				tenantUser, instanceID).Scan(&count)
		} else {
			err = s.db.QueryRowContext(ctx,
				`SELECT COUNT(*) FROM audit_logs WHERE tenant_user = $1 AND instance_id = $2 AND created_at >= $3`,
				tenantUser, instanceID, since).Scan(&count)
		}
	} else {
		if since.IsZero() {
			err = s.db.QueryRowContext(ctx,
				`SELECT COUNT(*) FROM audit_logs WHERE tenant_user = $1`,
				tenantUser).Scan(&count)
		} else {
			err = s.db.QueryRowContext(ctx,
				`SELECT COUNT(*) FROM audit_logs WHERE tenant_user = $1 AND created_at >= $2`,
				tenantUser, since).Scan(&count)
		}
	}
	return count, err
}

func (s *PostgresStore) countServiceLogsAll(ctx context.Context, tenantUser string, since time.Time, instanceID string) (int, error) {
	var count int
	var err error
	if instanceID != "" {
		if since.IsZero() {
			err = s.db.QueryRowContext(ctx,
				`SELECT COUNT(*) FROM service_logs WHERE tenant_user = $1 AND instance_id = $2`,
				tenantUser, instanceID).Scan(&count)
		} else {
			err = s.db.QueryRowContext(ctx,
				`SELECT COUNT(*) FROM service_logs WHERE tenant_user = $1 AND instance_id = $2 AND created_at >= $3`,
				tenantUser, instanceID, since).Scan(&count)
		}
	} else {
		if since.IsZero() {
			err = s.db.QueryRowContext(ctx,
				`SELECT COUNT(*) FROM service_logs WHERE tenant_user = $1`,
				tenantUser).Scan(&count)
		} else {
			err = s.db.QueryRowContext(ctx,
				`SELECT COUNT(*) FROM service_logs WHERE tenant_user = $1 AND created_at >= $2`,
				tenantUser, since).Scan(&count)
		}
	}
	return count, err
}

// ListLogsAll returns audit and/or service logs for the tenant, optionally filtered by instance, ordered by time desc.
func (s *PostgresStore) ListLogsAll(ctx context.Context, tenantUser string, opts ListOpts) ([]LogEntry, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	var entries []LogEntry

	switch opts.Type {
	case "audit":
		rows, err := s.queryAuditLogsAll(ctx, tenantUser, opts.Since, opts.InstanceID, limit, offset)
		if err != nil {
			return nil, err
		}
		entries = rows
	case "service":
		rows, err := s.queryServiceLogsAll(ctx, tenantUser, opts.Since, opts.InstanceID, limit, offset)
		if err != nil {
			return nil, err
		}
		entries = rows
	default:
		// Merged list with offset requires a single SQL query with UNION and OFFSET
		rows, err := s.queryMergedLogsAll(ctx, tenantUser, opts.Since, opts.InstanceID, limit, offset)
		if err != nil {
			return nil, err
		}
		entries = rows
	}

	return entries, nil
}

func (s *PostgresStore) queryAuditLogsAll(ctx context.Context, tenantUser string, since time.Time, instanceID string, limit, offset int) ([]LogEntry, error) {
	var rows *sql.Rows
	var err error
	if instanceID != "" {
		if since.IsZero() {
			rows, err = s.db.QueryContext(ctx,
				`SELECT id, instance_id, created_at, action, COALESCE(details::text, 'null') FROM audit_logs WHERE tenant_user = $1 AND instance_id = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
				tenantUser, instanceID, limit, offset)
		} else {
			rows, err = s.db.QueryContext(ctx,
				`SELECT id, instance_id, created_at, action, COALESCE(details::text, 'null') FROM audit_logs WHERE tenant_user = $1 AND instance_id = $2 AND created_at >= $3 ORDER BY created_at DESC LIMIT $4 OFFSET $5`,
				tenantUser, instanceID, since, limit, offset)
		}
	} else {
		if since.IsZero() {
			rows, err = s.db.QueryContext(ctx,
				`SELECT id, instance_id, created_at, action, COALESCE(details::text, 'null') FROM audit_logs WHERE tenant_user = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
				tenantUser, limit, offset)
		} else {
			rows, err = s.db.QueryContext(ctx,
				`SELECT id, instance_id, created_at, action, COALESCE(details::text, 'null') FROM audit_logs WHERE tenant_user = $1 AND created_at >= $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
				tenantUser, since, limit, offset)
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

func (s *PostgresStore) queryServiceLogsAll(ctx context.Context, tenantUser string, since time.Time, instanceID string, limit, offset int) ([]LogEntry, error) {
	var rows *sql.Rows
	var err error
	if instanceID != "" {
		if since.IsZero() {
			rows, err = s.db.QueryContext(ctx,
				`SELECT id, instance_id, created_at, event_type, message, COALESCE(metadata::text, 'null') FROM service_logs WHERE tenant_user = $1 AND instance_id = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
				tenantUser, instanceID, limit, offset)
		} else {
			rows, err = s.db.QueryContext(ctx,
				`SELECT id, instance_id, created_at, event_type, message, COALESCE(metadata::text, 'null') FROM service_logs WHERE tenant_user = $1 AND instance_id = $2 AND created_at >= $3 ORDER BY created_at DESC LIMIT $4 OFFSET $5`,
				tenantUser, instanceID, since, limit, offset)
		}
	} else {
		if since.IsZero() {
			rows, err = s.db.QueryContext(ctx,
				`SELECT id, instance_id, created_at, event_type, message, COALESCE(metadata::text, 'null') FROM service_logs WHERE tenant_user = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
				tenantUser, limit, offset)
		} else {
			rows, err = s.db.QueryContext(ctx,
				`SELECT id, instance_id, created_at, event_type, message, COALESCE(metadata::text, 'null') FROM service_logs WHERE tenant_user = $1 AND created_at >= $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
				tenantUser, since, limit, offset)
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

// queryMergedLogsAll returns audit and service logs merged by created_at desc with LIMIT and OFFSET.
func (s *PostgresStore) queryMergedLogsAll(ctx context.Context, tenantUser string, since time.Time, instanceID string, limit, offset int) ([]LogEntry, error) {
	// Build a UNION ALL query; same filters for both tables. Use subquery to apply ORDER BY LIMIT OFFSET once.
	auditSel := `SELECT id, instance_id, created_at, 'audit' AS log_type, action, NULL::text AS message, COALESCE(details::text, 'null') AS details, NULL::text AS metadata FROM audit_logs WHERE tenant_user = $1`
	serviceSel := `SELECT id, instance_id, created_at, 'service' AS log_type, event_type AS action, message, NULL::text AS details, COALESCE(metadata::text, 'null') AS metadata FROM service_logs WHERE tenant_user = $1`
	args := []any{tenantUser}
	argIdx := 2
	if instanceID != "" {
		auditSel += ` AND instance_id = $` + fmt.Sprintf("%d", argIdx)
		serviceSel += ` AND instance_id = $` + fmt.Sprintf("%d", argIdx)
		args = append(args, instanceID)
		argIdx++
	}
	if !since.IsZero() {
		auditSel += ` AND created_at >= $` + fmt.Sprintf("%d", argIdx)
		serviceSel += ` AND created_at >= $` + fmt.Sprintf("%d", argIdx)
		args = append(args, since)
		argIdx++
	}
	q := `SELECT id, instance_id, created_at, log_type, action, message, details, metadata FROM (` + auditSel + ` UNION ALL ` + serviceSel + `) AS u ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", argIdx) + ` OFFSET $` + fmt.Sprintf("%d", argIdx+1)
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []LogEntry
	for rows.Next() {
		var id, instID, logType, action string
		var createdAt time.Time
		var message, detailsNull, metadataNull sql.NullString
		if err := rows.Scan(&id, &instID, &createdAt, &logType, &action, &message, &detailsNull, &metadataNull); err != nil {
			return nil, err
		}
		e := LogEntry{
			ID:         id,
			Type:       logType,
			Timestamp:  createdAt,
			Action:     action,
			Message:    message.String,
			TenantUser: tenantUser,
			InstanceID: instID,
		}
		if logType == "audit" && detailsNull.Valid {
			e.Details = json.RawMessage(detailsNull.String)
		}
		if logType == "service" && metadataNull.Valid {
			e.Metadata = json.RawMessage(metadataNull.String)
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
