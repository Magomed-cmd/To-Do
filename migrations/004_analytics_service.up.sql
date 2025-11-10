CREATE SCHEMA IF NOT EXISTS analytics_service;

CREATE TABLE analytics_service.user_activity (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id INTEGER,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

CREATE TABLE analytics_service.task_metrics (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    date DATE NOT NULL,
    created_tasks INTEGER DEFAULT 0,
    completed_tasks INTEGER DEFAULT 0,
    total_tasks INTEGER DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, date)
);

CREATE TABLE analytics_service.export_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    export_type VARCHAR(20) NOT NULL CHECK (export_type IN ('csv', 'pdf', 'ical')),
    file_name VARCHAR(255),
    status VARCHAR(20) DEFAULT 'success',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_activity_user_id ON analytics_service.user_activity(user_id);