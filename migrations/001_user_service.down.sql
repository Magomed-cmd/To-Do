DROP TRIGGER IF EXISTS update_users_updated_at ON user_service.users;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS user_service.user_preferences;
DROP TABLE IF EXISTS user_service.user_sessions;
DROP TABLE IF EXISTS user_service.users;
DROP SCHEMA IF EXISTS user_service CASCADE;