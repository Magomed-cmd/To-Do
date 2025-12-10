DROP TRIGGER IF EXISTS update_tasks_updated_at ON task_service.tasks;
DROP TABLE IF EXISTS task_service.task_comments;
DROP TABLE IF EXISTS task_service.shared_list_members;
DROP TABLE IF EXISTS task_service.shared_lists;
DROP TABLE IF EXISTS task_service.tasks;
DROP TABLE IF EXISTS task_service.categories;
DROP SCHEMA IF EXISTS task_service CASCADE;
