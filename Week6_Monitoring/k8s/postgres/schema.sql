-- PaaS user-centric logs: audit (user actions) and service (async status/compliance).
-- Idempotent: safe to run multiple times (IF NOT EXISTS).
-- gen_random_uuid() is built-in in PostgreSQL 13+.

CREATE TABLE IF NOT EXISTS audit_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_user VARCHAR(255) NOT NULL,
  instance_id VARCHAR(255) NOT NULL,
  action VARCHAR(64) NOT NULL,
  details JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_instance_created
  ON audit_logs (tenant_user, instance_id, created_at DESC);

CREATE TABLE IF NOT EXISTS service_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_user VARCHAR(255) NOT NULL,
  instance_id VARCHAR(255) NOT NULL,
  event_type VARCHAR(128) NOT NULL,
  message TEXT,
  metadata JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_service_logs_tenant_instance_created
  ON service_logs (tenant_user, instance_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_service_logs_instance_created
  ON service_logs (instance_id, created_at DESC);
