DROP INDEX IF EXISTS uq_meeting_vote_options_key;

ALTER TABLE meeting_vote_records
    DROP COLUMN IF EXISTS weight,
    DROP COLUMN IF EXISTS option_key;

ALTER TABLE meeting_vote_options
    DROP COLUMN IF EXISTS option_key;

ALTER TABLE meeting_votes
    DROP COLUMN IF EXISTS updated_at;

ALTER TABLE meeting_attendees
    DROP COLUMN IF EXISTS checked_in_at;

ALTER TABLE meeting_agendas
    DROP COLUMN IF EXISTS presenter,
    DROP COLUMN IF EXISTS updated_at;

ALTER TABLE meetings
    DROP COLUMN IF EXISTS online_link,
    DROP COLUMN IF EXISTS minutes,
    DROP COLUMN IF EXISTS qr_token,
    DROP COLUMN IF EXISTS cancel_reason;

DELETE FROM permissions WHERE code = 'meeting:manage';
DELETE FROM configs WHERE config_key = 'vote_weight_config';
