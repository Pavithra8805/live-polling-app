package realtime

import "testing"

func TestHubBroadcastAndUnregister(t *testing.T) {
	hub := NewHub()
	first := hub.Register("poll-1")
	second := hub.Register("poll-1")
	hub.Broadcast("poll-1", []byte(`{"total":1}`))

	if got := string(<-first); got != `{"total":1}` {
		t.Fatalf("first payload = %q", got)
	}
	if got := string(<-second); got != `{"total":1}` {
		t.Fatalf("second payload = %q", got)
	}

	hub.Unregister("poll-1", first)
	hub.Unregister("poll-1", second)
	hub.Broadcast("poll-1", []byte(`{"total":2}`))
}
