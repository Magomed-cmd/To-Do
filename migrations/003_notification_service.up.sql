CREATE SCHEMA IF NOT EXISTS notification_service;

CREATE TABLE notification_service.notification_templates (
    id SERIAL PRIMARY KEY,
    type VARCHAR(50) UNIQUE NOT NULL,
    subject VARCHAR(255) NOT NULL,
    body_template TEXT NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE notification_service.notification_queue (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    type VARCHAR(50) NOT NULL,
    subject VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    payload JSONB,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'failed')),
    scheduled_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    sent_at TIMESTAMP,
    attempts INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE notification_service.user_notification_settings (
    user_id INTEGER PRIMARY KEY,
    email_enabled BOOLEAN DEFAULT true,
    task_due_reminder BOOLEAN DEFAULT true,
    task_assigned_notification BOOLEAN DEFAULT true,
    reminder_intervals INTEGER[] DEFAULT ARRAY[60, 1440],
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notification_queue_user_id ON notification_service.notification_queue(user_id);
CREATE INDEX idx_notification_queue_status ON notification_service.notification_queue(status);