-- ============================================================
-- 000012_meeting_tables.up.sql
-- 会议、议程、参会人、投票（Issue #19 原 000005）
-- ============================================================

CREATE TABLE meetings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(200) NOT NULL,
    description TEXT,
    status SMALLINT NOT NULL DEFAULT 0, -- 0=草稿 1=已发布 2=进行中 3=已结束
    meeting_type SMALLINT NOT NULL DEFAULT 1, -- 1=例会 2=临时会议 3=线上
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    location VARCHAR(200),
    organizer_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_meetings_status ON meetings(status);
CREATE INDEX idx_meetings_organizer_id ON meetings(organizer_id);
CREATE INDEX idx_meetings_start_time ON meetings(start_time);
CREATE INDEX idx_meetings_created_at ON meetings(created_at);

CREATE TABLE meeting_agendas (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    meeting_id UUID NOT NULL REFERENCES meetings(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    sort_order INT NOT NULL DEFAULT 0,
    duration INT,
    speaker_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_meeting_agendas_meeting_id ON meeting_agendas(meeting_id);

CREATE TABLE meeting_attendees (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    meeting_id UUID NOT NULL REFERENCES meetings(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    attended BOOLEAN DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (meeting_id, user_id)
);

CREATE INDEX idx_meeting_attendees_user_id ON meeting_attendees(user_id);

CREATE TABLE meeting_votes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    meeting_id UUID NOT NULL REFERENCES meetings(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    vote_type SMALLINT NOT NULL DEFAULT 1, -- 1=等权 2=加权
    status SMALLINT NOT NULL DEFAULT 0, -- 0=未开始 1=进行中 2=已结束
    is_anonymous BOOLEAN NOT NULL DEFAULT false,
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_meeting_votes_meeting_id ON meeting_votes(meeting_id);
CREATE INDEX idx_meeting_votes_status ON meeting_votes(status);

CREATE TABLE meeting_vote_options (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    vote_id UUID NOT NULL REFERENCES meeting_votes(id) ON DELETE CASCADE,
    option_text VARCHAR(200) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0
);

CREATE INDEX idx_meeting_vote_options_vote_id ON meeting_vote_options(vote_id);

CREATE TABLE meeting_vote_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    vote_id UUID NOT NULL REFERENCES meeting_votes(id) ON DELETE CASCADE,
    voter_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    option_id UUID NOT NULL REFERENCES meeting_vote_options(id) ON DELETE RESTRICT,
    voted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (vote_id, voter_id)
);

CREATE INDEX idx_meeting_vote_records_vote_id ON meeting_vote_records(vote_id);
CREATE INDEX idx_meeting_vote_records_voter_id ON meeting_vote_records(voter_id);
