package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"terraform-provider-wordpress/internal/wpapi"
)

func TestApplicationPasswordEphemeralDeleteOnCloseHelper(t *testing.T) {
	var sawDelete bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodDelete && r.URL.Path == "/wp-json/wp/v2/users/1/application-passwords/uuid-1":
			sawDelete = true
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"deleted":true}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := wpapi.New(server.URL+"/wp-json/wp/v2", "admin", "secret")
	if err != nil {
		t.Fatalf("wpapi.New returned error: %v", err)
	}

	r := &applicationPasswordEphemeralResource{client: client}
	if err := r.deleteApplicationPasswordIfRequested(context.Background(), 1, "uuid-1", false); err != nil {
		t.Fatalf("unexpected error with delete_on_close=false: %v", err)
	}
	if sawDelete {
		t.Fatal("expected delete_on_close=false to skip deletion")
	}

	if err := r.deleteApplicationPasswordIfRequested(context.Background(), 1, "uuid-1", true); err != nil {
		t.Fatalf("unexpected error with delete_on_close=true: %v", err)
	}
	if !sawDelete {
		t.Fatal("expected delete_on_close=true to delete the password")
	}
}

func TestApplicationPasswordEphemeralCleanupStateJSON(t *testing.T) {
	state := applicationPasswordEphemeralCleanupState{
		DeleteOnClose: true,
		UserID:        1,
		UUID:          "uuid-1",
	}

	encoded, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded applicationPasswordEphemeralCleanupState
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.UserID != state.UserID || decoded.UUID != state.UUID || !decoded.DeleteOnClose {
		t.Fatalf("unexpected decoded state: %#v", decoded)
	}
}
