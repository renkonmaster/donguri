package handler

import "testing"

func TestParseRoomStreamPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		path   string
		wantID string
		ok     bool
	}{
		{name: "valid", path: "/api/rooms/550e8400-e29b-41d4-a716-446655440000/stream", wantID: "550e8400-e29b-41d4-a716-446655440000", ok: true},
		{name: "invalid prefix", path: "/api/room/abc/stream", ok: false},
		{name: "invalid suffix", path: "/api/rooms/abc/messages", ok: false},
		{name: "missing room id", path: "/api/rooms//stream", ok: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotID, gotOK := parseRoomStreamPath(tc.path)
			if gotOK != tc.ok {
				t.Fatalf("unexpected ok: got=%v want=%v", gotOK, tc.ok)
			}
			if gotID != tc.wantID {
				t.Fatalf("unexpected room id: got=%q want=%q", gotID, tc.wantID)
			}
		})
	}
}
