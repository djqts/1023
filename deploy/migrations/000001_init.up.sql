CREATE TABLE IF NOT EXISTS rule_set (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  version VARCHAR(64) NOT NULL,
  digest CHAR(64) NOT NULL,
  content LONGTEXT NOT NULL,
  source VARCHAR(32) NOT NULL,
  status VARCHAR(16) NOT NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  activated_at DATETIME(6) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_rule_set_version (version),
  KEY idx_rule_set_digest (digest),
  KEY idx_rule_set_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS recognition_batch (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  request_id VARCHAR(64) NOT NULL,
  item_count INT NOT NULL,
  unknown_count INT NOT NULL,
  latency_ms BIGINT NOT NULL,
  rule_digest CHAR(64) NOT NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uk_recognition_batch_request (request_id),
  KEY idx_recognition_batch_rule (rule_digest),
  KEY idx_recognition_batch_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS recognition_item (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  batch_id BIGINT UNSIGNED NOT NULL,
  ip_hash CHAR(64) NOT NULL,
  port INT NOT NULL,
  protocol VARCHAR(16) NOT NULL,
  product VARCHAR(64) NOT NULL,
  version VARCHAR(64) NOT NULL,
  os_hint VARCHAR(64) NOT NULL,
  confidence DOUBLE NOT NULL,
  rule_id VARCHAR(64) NOT NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  KEY idx_recognition_item_batch (batch_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS async_job (
  id VARCHAR(64) NOT NULL,
  idempotency_key VARCHAR(128) NOT NULL,
  status VARCHAR(16) NOT NULL,
  input_ref VARCHAR(256) NOT NULL,
  result_ref VARCHAR(256) NOT NULL,
  error_code VARCHAR(64) NOT NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uk_async_job_idem (idempotency_key),
  KEY idx_async_job_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS outbox_event (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  aggregate_type VARCHAR(32) NOT NULL,
  aggregate_id VARCHAR(64) NOT NULL,
  event_type VARCHAR(32) NOT NULL,
  payload JSON NOT NULL,
  status VARCHAR(16) NOT NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  published_at DATETIME(6) NULL,
  PRIMARY KEY (id),
  KEY idx_outbox_aggregate (aggregate_type, aggregate_id),
  KEY idx_outbox_status_created (status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
