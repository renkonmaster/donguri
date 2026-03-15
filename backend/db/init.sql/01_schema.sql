CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE IF NOT EXISTS rooms (
	id UUID PRIMARY KEY,
	status TEXT NOT NULL CHECK (status IN ('matching', 'playing', 'finished')),
	start_at TIMESTAMPTZ,
	expires_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS players (
	id UUID PRIMARY KEY,
	room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
	name VARCHAR(255) NOT NULL,
	location geography(Point, 4326) NOT NULL,
	order_index INT NOT NULL CHECK (order_index >= 0),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	UNIQUE (room_id, order_index)
);
CREATE INDEX IF NOT EXISTS idx_players_room_id ON players(room_id);

CREATE TABLE IF NOT EXISTS messages (
	id UUID PRIMARY KEY,
	room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
	sender_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
	receiver_id UUID REFERENCES players(id) ON DELETE CASCADE,
	content TEXT NOT NULL CHECK (char_length(content) > 0),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_messages_room_id ON messages(room_id);
CREATE INDEX IF NOT EXISTS idx_messages_sender_id ON messages(sender_id);
CREATE INDEX IF NOT EXISTS idx_messages_receiver_id ON messages(receiver_id);

CREATE TABLE IF NOT EXISTS connections (
	room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
	sender_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
	receiver_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
	needs_swap BOOLEAN NOT NULL DEFAULT FALSE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	PRIMARY KEY (room_id, sender_id, receiver_id),
	CHECK (sender_id <> receiver_id)
);
CREATE INDEX IF NOT EXISTS idx_connections_sender_id ON connections(sender_id);
CREATE INDEX IF NOT EXISTS idx_connections_receiver_id ON connections(receiver_id);

CREATE OR REPLACE FUNCTION get_room_intersection_count(target_room_id UUID)
RETURNS INTEGER
LANGUAGE SQL
STABLE
AS $$
WITH ordered_players AS (
	SELECT
		id,
		order_index,
		location::geometry AS geom,
		LEAD(id) OVER (ORDER BY order_index) AS next_id,
		LEAD(location::geometry) OVER (ORDER BY order_index) AS next_geom,
		FIRST_VALUE(id) OVER (ORDER BY order_index) AS first_id,
		FIRST_VALUE(location::geometry) OVER (ORDER BY order_index) AS first_geom
	FROM players
	WHERE room_id = target_room_id
),
segments AS (
	SELECT
		id AS from_id,
		COALESCE(next_id, first_id) AS to_id,
		ST_MakeLine(geom, COALESCE(next_geom, first_geom)) AS segment
	FROM ordered_players
),
segment_pairs AS (
	SELECT
		s1.segment AS segment1,
		s2.segment AS segment2
	FROM segments s1
	JOIN segments s2 ON s1.from_id < s2.from_id
	WHERE
		s1.from_id <> s2.from_id
		AND s1.from_id <> s2.to_id
		AND s1.to_id <> s2.from_id
		AND s1.to_id <> s2.to_id
)
SELECT COALESCE(COUNT(*), 0)::INTEGER
FROM segment_pairs
WHERE ST_Crosses(segment1, segment2);
$$;