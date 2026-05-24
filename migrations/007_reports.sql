CREATE TABLE IF NOT EXISTS data_insight.report_tasks (
  id uuid PRIMARY KEY,
  tenant_id text NOT NULL,
  report_type text NOT NULL,
  status text NOT NULL CHECK (status IN ('pending', 'running', 'success', 'failed')),
  format text NOT NULL DEFAULT 'xlsx',
  parameters jsonb NOT NULL DEFAULT '{}'::jsonb,
  file_name text,
  file_ref text,
  download_token_hash text,
  row_count int NOT NULL DEFAULT 0,
  summary jsonb NOT NULL DEFAULT '{}'::jsonb,
  consistency jsonb NOT NULL DEFAULT '{}'::jsonb,
  advice jsonb NOT NULL DEFAULT '[]'::jsonb,
  created_by text,
  created_at timestamptz NOT NULL DEFAULT now(),
  completed_at timestamptz
);

CREATE TABLE IF NOT EXISTS data_insight.report_download_logs (
  id uuid PRIMARY KEY,
  report_task_id uuid NOT NULL REFERENCES data_insight.report_tasks(id),
  tenant_id text NOT NULL,
  downloaded_by text,
  token_ref text,
  status text NOT NULL CHECK (status IN ('success', 'denied', 'expired')),
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS data_insight.report_action_advice (
  id uuid PRIMARY KEY,
  tenant_id text NOT NULL,
  report_task_id uuid REFERENCES data_insight.report_tasks(id),
  severity text NOT NULL CHECK (severity IN ('low', 'medium', 'high')),
  metric_code text NOT NULL,
  title text NOT NULL,
  detail text NOT NULL,
  action text NOT NULL,
  status text NOT NULL DEFAULT 'open',
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_report_tasks_tenant_created ON data_insight.report_tasks(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_report_download_logs_task ON data_insight.report_download_logs(report_task_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_report_action_advice_tenant_status ON data_insight.report_action_advice(tenant_id, status);
