CREATE INDEX idx_results_monitor_checked_at
ON results (monitor_id, checked_at DESC);