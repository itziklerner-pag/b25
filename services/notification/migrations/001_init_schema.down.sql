-- Rollback notification service schema

-- Drop views
DROP VIEW IF EXISTS user_notification_summary;
DROP VIEW IF EXISTS notification_stats_by_channel;

-- Drop triggers
DROP TRIGGER IF EXISTS update_alert_rules_updated_at ON alert_rules;
DROP TRIGGER IF EXISTS update_webhooks_updated_at ON webhooks;
DROP TRIGGER IF EXISTS update_notifications_updated_at ON notifications;
DROP TRIGGER IF EXISTS update_notification_templates_updated_at ON notification_templates;
DROP TRIGGER IF EXISTS update_notification_preferences_updated_at ON notification_preferences;
DROP TRIGGER IF EXISTS update_user_devices_updated_at ON user_devices;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse order
DROP TABLE IF EXISTS alert_rule_executions;
DROP TABLE IF EXISTS alert_rules;
DROP TABLE IF EXISTS webhook_deliveries;
DROP TABLE IF EXISTS webhooks;
DROP TABLE IF EXISTS rate_limits;
DROP TABLE IF EXISTS notification_batches;
DROP TABLE IF EXISTS notification_events;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS notification_templates;
DROP TABLE IF EXISTS notification_preferences;
DROP TABLE IF EXISTS user_devices;
DROP TABLE IF EXISTS users;

-- Drop types
DROP TYPE IF EXISTS template_type;
DROP TYPE IF EXISTS notification_priority;
DROP TYPE IF EXISTS notification_status;
DROP TYPE IF EXISTS notification_channel;
