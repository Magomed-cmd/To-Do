CREATE SCHEMA IF NOT EXISTS task_service;

CREATE TABLE task_service.categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7) DEFAULT '#3498db',
    user_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(name, user_id)
);

CREATE TABLE task_service.tasks (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'completed', 'archived')),
    priority VARCHAR(10) DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high')),
    due_date TIMESTAMP,
    completed_at TIMESTAMP,
    user_id INTEGER NOT NULL,
    assigned_to INTEGER,
    category_id INTEGER REFERENCES task_service.categories(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE task_service.shared_lists (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    owner_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE task_service.shared_list_members (
    list_id INTEGER REFERENCES task_service.shared_lists(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL,
    permission_level VARCHAR(20) DEFAULT 'viewer' CHECK (permission_level IN ('viewer', 'editor', 'admin')),
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (list_id, user_id)
);

CREATE TABLE task_service.task_comments (
    id SERIAL PRIMARY KEY,
    task_id INTEGER REFERENCES task_service.tasks(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tasks_user_id ON task_service.tasks(user_id);
CREATE INDEX idx_tasks_status ON task_service.tasks(status);
CREATE INDEX idx_tasks_due_date ON task_service.tasks(due_date);

CREATE TRIGGER update_tasks_updated_at BEFORE UPDATE ON task_service.tasks FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
