package wpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientPagesCRUD(t *testing.T) {
	var sawList bool
	var sawCreate bool
	var sawUpdate bool
	var sawDelete bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/wp-json/wp/v2/pages":
			if got := r.URL.Query().Get("context"); got != "edit" {
				t.Fatalf("unexpected list context: %q", got)
			}
			if got := r.URL.Query().Get("per_page"); got != "100" {
				t.Fatalf("unexpected list per_page: %q", got)
			}
			if user, pass, ok := r.BasicAuth(); !ok || user != "admin" || pass != "secret" {
				t.Fatalf("missing basic auth: %q %q %v", user, pass, ok)
			}
			sawList = true
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]Page{{ID: 1, Title: RenderedField{Rendered: "Hello", Raw: "Hello"}, Content: ContentField{Rendered: "Body", Raw: "Body"}, Excerpt: ProtectedField{Rendered: "Summary", Raw: "Summary"}}})
		case r.Method == http.MethodPost && r.URL.Path == "/wp-json/wp/v2/pages/":
			if user, pass, ok := r.BasicAuth(); !ok || user != "admin" || pass != "secret" {
				t.Fatalf("missing basic auth on create: %q %q %v", user, pass, ok)
			}
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode create payload: %v", err)
			}
			if payload["title"] != "Created" || payload["content"] != "Body" || payload["excerpt"] != "Summary" {
				t.Fatalf("unexpected create payload: %#v", payload)
			}
			sawCreate = true
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(Page{ID: 2, Title: RenderedField{Rendered: "Created", Raw: "Created"}, Content: ContentField{Rendered: "Body", Raw: "Body"}, Excerpt: ProtectedField{Rendered: "Summary", Raw: "Summary"}})
		case r.Method == http.MethodGet && r.URL.Path == "/wp-json/wp/v2/pages/2":
			if got := r.URL.Query().Get("context"); got != "edit" {
				t.Fatalf("unexpected get context: %q", got)
			}
			sawCreate = sawCreate && true
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(Page{ID: 2, Title: RenderedField{Rendered: "Created", Raw: "Created"}, Content: ContentField{Rendered: "Body", Raw: "Body"}, Excerpt: ProtectedField{Rendered: "Summary", Raw: "Summary"}})
		case r.Method == http.MethodPost && r.URL.Path == "/wp-json/wp/v2/pages/2":
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode update payload: %v", err)
			}
			if payload["title"] != "Updated" {
				t.Fatalf("unexpected update payload: %#v", payload)
			}
			sawUpdate = true
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(Page{ID: 2, Title: RenderedField{Rendered: "Updated", Raw: "Updated"}, Content: ContentField{Rendered: "Body", Raw: "Body"}, Excerpt: ProtectedField{Rendered: "Summary", Raw: "Summary"}})
		case r.Method == http.MethodDelete && r.URL.Path == "/wp-json/wp/v2/pages/2":
			if got := r.URL.Query().Get("force"); got != "true" {
				t.Fatalf("unexpected delete force: %q", got)
			}
			sawDelete = true
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"deleted":true}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := New(server.URL+"/wp-json/wp/v2", "admin", "secret")
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	pages, err := client.ListPages(context.Background())
	if err != nil {
		t.Fatalf("ListPages returned error: %v", err)
	}
	if len(pages) != 1 || pages[0].ID != 1 {
		t.Fatalf("unexpected list result: %#v", pages)
	}

	created, err := client.CreatePage(context.Background(), PageInput{Title: stringPtr("Created"), Content: stringPtr("Body"), Excerpt: stringPtr("Summary")})
	if err != nil {
		t.Fatalf("CreatePage returned error: %v", err)
	}
	if created.ID != 2 || created.Title.Rendered != "Created" {
		t.Fatalf("unexpected create result: %#v", created)
	}

	page, err := client.GetPage(context.Background(), 2)
	if err != nil {
		t.Fatalf("GetPage returned error: %v", err)
	}
	if page.ID != 2 {
		t.Fatalf("unexpected get result: %#v", page)
	}

	updated, err := client.UpdatePage(context.Background(), 2, PageInput{Title: stringPtr("Updated")})
	if err != nil {
		t.Fatalf("UpdatePage returned error: %v", err)
	}
	if updated.Title.Rendered != "Updated" {
		t.Fatalf("unexpected update result: %#v", updated)
	}

	if err := client.DeletePage(context.Background(), 2); err != nil {
		t.Fatalf("DeletePage returned error: %v", err)
	}

	if !sawList || !sawCreate || !sawUpdate || !sawDelete {
		t.Fatalf("missing calls: list=%v create=%v update=%v delete=%v", sawList, sawCreate, sawUpdate, sawDelete)
	}
}

func TestClientPostsCRUD(t *testing.T) {
	var sawList bool
	var sawCreate bool
	var sawGet bool
	var sawUpdate bool
	var sawDelete bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/wp-json/wp/v2/posts":
			if got := r.URL.Query().Get("context"); got != "edit" {
				t.Fatalf("unexpected list context: %q", got)
			}
			if got := r.URL.Query().Get("per_page"); got != "100" {
				t.Fatalf("unexpected list per_page: %q", got)
			}
			if user, pass, ok := r.BasicAuth(); !ok || user != "admin" || pass != "secret" {
				t.Fatalf("missing basic auth: %q %q %v", user, pass, ok)
			}
			sawList = true
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]Post{{ID: 1, Title: RenderedField{Rendered: "Hello", Raw: "Hello"}, Content: ContentField{Rendered: "Body", Raw: "Body"}, Excerpt: ProtectedField{Rendered: "Summary", Raw: "Summary"}, Sticky: true}})
		case r.Method == http.MethodPost && r.URL.Path == "/wp-json/wp/v2/posts/":
			if user, pass, ok := r.BasicAuth(); !ok || user != "admin" || pass != "secret" {
				t.Fatalf("missing basic auth on create: %q %q %v", user, pass, ok)
			}
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode create payload: %v", err)
			}
			if payload["title"] != "Created" || payload["content"] != "Body" || payload["excerpt"] != "Summary" {
				t.Fatalf("unexpected create payload: %#v", payload)
			}
			if payload["sticky"] != true || payload["format"] != "standard" {
				t.Fatalf("unexpected create flags: %#v", payload)
			}
			sawCreate = true
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(Post{ID: 2, Title: RenderedField{Rendered: "Created", Raw: "Created"}, Content: ContentField{Rendered: "Body", Raw: "Body"}, Excerpt: ProtectedField{Rendered: "Summary", Raw: "Summary"}, Sticky: true, Format: "standard"})
		case r.Method == http.MethodGet && r.URL.Path == "/wp-json/wp/v2/posts/2":
			if got := r.URL.Query().Get("context"); got != "edit" {
				t.Fatalf("unexpected get context: %q", got)
			}
			sawGet = true
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(Post{ID: 2, Title: RenderedField{Rendered: "Created", Raw: "Created"}, Content: ContentField{Rendered: "Body", Raw: "Body"}, Excerpt: ProtectedField{Rendered: "Summary", Raw: "Summary"}, Sticky: true, Format: "standard"})
		case r.Method == http.MethodPost && r.URL.Path == "/wp-json/wp/v2/posts/2":
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode update payload: %v", err)
			}
			if payload["title"] != "Updated" {
				t.Fatalf("unexpected update payload: %#v", payload)
			}
			sawUpdate = true
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(Post{ID: 2, Title: RenderedField{Rendered: "Updated", Raw: "Updated"}, Content: ContentField{Rendered: "Body", Raw: "Body"}, Excerpt: ProtectedField{Rendered: "Summary", Raw: "Summary"}, Sticky: false, Format: "aside"})
		case r.Method == http.MethodDelete && r.URL.Path == "/wp-json/wp/v2/posts/2":
			if got := r.URL.Query().Get("force"); got != "true" {
				t.Fatalf("unexpected delete force: %q", got)
			}
			sawDelete = true
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"deleted":true}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := New(server.URL+"/wp-json/wp/v2", "admin", "secret")
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	posts, err := client.ListPosts(context.Background())
	if err != nil {
		t.Fatalf("ListPosts returned error: %v", err)
	}
	if len(posts) != 1 || posts[0].ID != 1 {
		t.Fatalf("unexpected list result: %#v", posts)
	}

	created, err := client.CreatePost(context.Background(), PostInput{Title: stringPtr("Created"), Content: stringPtr("Body"), Excerpt: stringPtr("Summary"), Sticky: boolPtr(true), Format: stringPtr("standard")})
	if err != nil {
		t.Fatalf("CreatePost returned error: %v", err)
	}
	if created.ID != 2 || created.Title.Rendered != "Created" {
		t.Fatalf("unexpected create result: %#v", created)
	}

	post, err := client.GetPost(context.Background(), 2)
	if err != nil {
		t.Fatalf("GetPost returned error: %v", err)
	}
	if post.ID != 2 {
		t.Fatalf("unexpected get result: %#v", post)
	}

	updated, err := client.UpdatePost(context.Background(), 2, PostInput{Title: stringPtr("Updated")})
	if err != nil {
		t.Fatalf("UpdatePost returned error: %v", err)
	}
	if updated.Title.Rendered != "Updated" {
		t.Fatalf("unexpected update result: %#v", updated)
	}

	if err := client.DeletePost(context.Background(), 2); err != nil {
		t.Fatalf("DeletePost returned error: %v", err)
	}

	if !sawList || !sawCreate || !sawGet || !sawUpdate || !sawDelete {
		t.Fatalf("missing calls: list=%v create=%v get=%v update=%v delete=%v", sawList, sawCreate, sawGet, sawUpdate, sawDelete)
	}
}

func TestClientUsersCRUD(t *testing.T) {
	var sawList bool
	var sawCreate bool
	var sawGet bool
	var sawUpdate bool
	var sawDelete bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/wp-json/wp/v2/users":
			if got := r.URL.Query().Get("context"); got != "edit" {
				t.Fatalf("unexpected list context: %q", got)
			}
			if got := r.URL.Query().Get("per_page"); got != "100" {
				t.Fatalf("unexpected list per_page: %q", got)
			}
			if user, pass, ok := r.BasicAuth(); !ok || user != "admin" || pass != "secret" {
				t.Fatalf("missing basic auth: %q %q %v", user, pass, ok)
			}
			sawList = true
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]User{{ID: 1, Username: "admin", Name: "Admin", Email: "admin@example.com", Roles: []string{"administrator"}}})
		case r.Method == http.MethodPost && r.URL.Path == "/wp-json/wp/v2/users/":
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode create payload: %v", err)
			}
			if payload["username"] != "newuser" || payload["email"] != "newuser@example.com" || payload["password"] != "secret" {
				t.Fatalf("unexpected create payload: %#v", payload)
			}
			roles, ok := payload["roles"].([]any)
			if !ok || len(roles) != 1 || roles[0] != "author" {
				t.Fatalf("unexpected create roles: %#v", payload["roles"])
			}
			sawCreate = true
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(User{ID: 2, Username: "newuser", Name: "New User", Email: "newuser@example.com", Roles: []string{"author"}})
		case r.Method == http.MethodGet && r.URL.Path == "/wp-json/wp/v2/users/2":
			if got := r.URL.Query().Get("context"); got != "edit" {
				t.Fatalf("unexpected get context: %q", got)
			}
			sawGet = true
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(User{ID: 2, Username: "newuser", Name: "New User", Email: "newuser@example.com", Roles: []string{"author"}})
		case r.Method == http.MethodPost && r.URL.Path == "/wp-json/wp/v2/users/2":
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode update payload: %v", err)
			}
			if payload["name"] != "Updated User" {
				t.Fatalf("unexpected update payload: %#v", payload)
			}
			sawUpdate = true
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(User{ID: 2, Username: "newuser", Name: "Updated User", Email: "newuser@example.com", Roles: []string{"author"}})
		case r.Method == http.MethodDelete && r.URL.Path == "/wp-json/wp/v2/users/2":
			if got := r.URL.Query().Get("force"); got != "true" {
				t.Fatalf("unexpected delete force: %q", got)
			}
			if got := r.URL.Query().Get("reassign"); got != "1" {
				t.Fatalf("unexpected delete reassign: %q", got)
			}
			sawDelete = true
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"deleted":true}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := New(server.URL+"/wp-json/wp/v2", "admin", "secret")
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	users, err := client.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("ListUsers returned error: %v", err)
	}
	if len(users) != 1 || users[0].ID != 1 {
		t.Fatalf("unexpected list result: %#v", users)
	}

	created, err := client.CreateUser(context.Background(), UserInput{
		Username: stringPtr("newuser"),
		Email:    stringPtr("newuser@example.com"),
		Password: stringPtr("secret"),
		Roles:    []string{"author"},
	})
	if err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}
	if created.ID != 2 || created.Username != "newuser" {
		t.Fatalf("unexpected create result: %#v", created)
	}

	user, err := client.GetUser(context.Background(), 2)
	if err != nil {
		t.Fatalf("GetUser returned error: %v", err)
	}
	if user.ID != 2 {
		t.Fatalf("unexpected get result: %#v", user)
	}

	updated, err := client.UpdateUser(context.Background(), 2, UserInput{Name: stringPtr("Updated User")})
	if err != nil {
		t.Fatalf("UpdateUser returned error: %v", err)
	}
	if updated.Name != "Updated User" {
		t.Fatalf("unexpected update result: %#v", updated)
	}

	if err := client.DeleteUser(context.Background(), 2, 1); err != nil {
		t.Fatalf("DeleteUser returned error: %v", err)
	}

	if !sawList || !sawCreate || !sawGet || !sawUpdate || !sawDelete {
		t.Fatalf("missing calls: list=%v create=%v get=%v update=%v delete=%v", sawList, sawCreate, sawGet, sawUpdate, sawDelete)
	}
}

func TestNewRequiresBaseURL(t *testing.T) {
	if _, err := New("", "admin", "secret"); err == nil {
		t.Fatal("expected error for empty base URL")
	}
}

func stringPtr(value string) *string {
	return &value
}

func boolPtr(value bool) *bool {
	return &value
}
