-- 000051_schedule_calendar.down.sql

UPDATE scheduler_tasks SET status = 2, updated_at = NOW()
WHERE code = 'schedule_reminder' AND status <> 2;

DELETE FROM permissions WHERE code IN (
    'schedule:read', 'schedule:create', 'schedule:update', 'schedule:delete'
);

DROP TABLE IF EXISTS schedule_reminders;
DROP TABLE IF EXISTS schedule_event_attendees;
DROP TABLE IF EXISTS schedule_events;
DROP TABLE IF EXISTS calendar_members;
DROP TABLE IF EXISTS calendars;
