-- +goose Up
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

-- +goose Down
DROP TABLE IF EXISTS connections;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS players;
DROP TABLE IF EXISTS rooms;
