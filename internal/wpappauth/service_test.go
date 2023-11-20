package wpappauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestCreateApplicationPassword(t *testing.T) {
	var sawLoginCookie bool
	var sawPermalinkUpdate bool
	var sawNonceHeader bool
	var restJSONCheckedBeforePermalink bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/wp-login.php":
			http.SetCookie(w, &http.Cookie{Name: "wordpress_test_cookie", Value: "WP Cookie check", Path: "/"})
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("login page"))
		case r.Method == http.MethodPost && r.URL.Path == "/wp-login.php":
			if got := r.PostFormValue("log"); got != "admin" {
				t.Fatalf("unexpected username: %q", got)
			}
			if got := r.PostFormValue("pwd"); got != "secret" {
				t.Fatalf("unexpected password: %q", got)
			}
			if got := r.PostFormValue("testcookie"); got != "1" {
				t.Fatalf("unexpected testcookie: %q", got)
			}
			if cookie := r.Header.Get("Cookie"); !strings.Contains(cookie, "wordpress_test_cookie=") {
				t.Fatalf("login request did not send test cookie: %q", cookie)
			}
			http.SetCookie(w, &http.Cookie{Name: "wordpress_logged_in_test", Value: "admin|session", Path: "/"})
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("logged in"))
		case r.Method == http.MethodGet && r.URL.Path == "/wp-admin/profile.php":
			if cookie := r.Header.Get("Cookie"); strings.Contains(cookie, "wordpress_logged_in_test=admin|session") {
				sawLoginCookie = true
			}
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintf(w, `<html><script>var userId = 1;</script><script>var wpApiSettings = {"root":"http://%s/index.php?rest_route=/","nonce":"abc123","versionString":"wp/v2/"};</script></html>`, r.Host)
		case r.Method == http.MethodGet && r.URL.Path == "/wp-json":
			if !sawPermalinkUpdate {
				restJSONCheckedBeforePermalink = true
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte("not json"))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"name":"wordpress"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/wp-admin/options-permalink.php":
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintf(w, `<html><form><input type="hidden" name="_wpnonce" value="permalink-nonce" /><input type="hidden" name="_wp_http_referer" value="/wp-admin/options-permalink.php" /></form></html>`)
		case r.Method == http.MethodPost && r.URL.Path == "/wp-admin/options-permalink.php":
			if got := r.PostFormValue("_wpnonce"); got != "permalink-nonce" {
				t.Fatalf("unexpected permalink nonce: %q", got)
			}
			if got := r.PostFormValue("permalink_structure"); got != "/%year%/%monthnum%/%day%/%postname%/" {
				t.Fatalf("unexpected permalink structure: %q", got)
			}
			sawPermalinkUpdate = true
			w.WriteHeader(http.StatusFound)
		case r.Method == http.MethodPost && r.URL.Path == "/index.php":
			if got := r.URL.Query().Get("rest_route"); got != "/wp/v2/users/1/application-passwords?_locale=user" {
				t.Fatalf("unexpected rest_route: %q", got)
			}
			if got := r.Header.Get("X-WP-Nonce"); got != "abc123" {
				t.Fatalf("unexpected nonce header: %q", got)
			}
			if got := r.PostFormValue("name"); got != "terraform-provider-wordpress" {
				t.Fatalf("unexpected application name: %q", got)
			}
			if cookie := r.Header.Get("Cookie"); strings.Contains(cookie, "wordpress_logged_in_test=admin|session") {
				sawNonceHeader = true
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(Result{UUID: "uuid-1", AppID: "app-1", Name: "terraform-provider-wordpress", Password: "abcd efgh ijkl mnop"})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	service := Service{
		BaseURL:  server.URL + "/wp-json/wp/v2",
		Username: "admin",
		Password: "secret",
	}

	result, err := service.CreateApplicationPassword(context.Background())
	if err != nil {
		t.Fatalf("CreateApplicationPassword returned error: %v", err)
	}
	if result.Password != "abcd efgh ijkl mnop" {
		t.Fatalf("unexpected password: %q", result.Password)
	}
	if !sawLoginCookie {
		t.Fatal("profile request did not carry logged-in cookie")
	}
	if !restJSONCheckedBeforePermalink {
		t.Fatal("rest api was not checked before permalink repair")
	}
	if !sawPermalinkUpdate {
		t.Fatal("permalink update was not attempted when wp-json was invalid")
	}
	if !sawNonceHeader {
		t.Fatal("application password request did not carry logged-in cookie")
	}
}

func TestSiteBaseURL(t *testing.T) {
	parsed, err := siteBaseURL("http://example.com/wp-json/wp/v2")
	if err != nil {
		t.Fatalf("siteBaseURL returned error: %v", err)
	}
	if got, want := parsed.String(), "http://example.com/"; got != want {
		t.Fatalf("unexpected site base URL: got %q want %q", got, want)
	}
}

func TestCreateApplicationPasswordRequiresFields(t *testing.T) {
	service := Service{}
	if _, err := service.CreateApplicationPassword(context.Background()); err == nil {
		t.Fatal("expected error for missing fields")
	}
}

func TestJoinPath(t *testing.T) {
	base, err := url.Parse("http://example.com/")
	if err != nil {
		t.Fatalf("failed to parse base URL: %v", err)
	}
	if got, want := joinPath(base, "wp-login.php"), "http://example.com/wp-login.php"; got != want {
		t.Fatalf("unexpected joinPath result: got %q want %q", got, want)
	}
}

func TestRestRouteURL(t *testing.T) {
	if got, want := restRouteURL("http://example.test/index.php?rest_route=/", "/wp/v2/users/1/application-passwords?_locale=user"), "http://example.test/index.php?rest_route=/wp/v2/users/1/application-passwords?_locale=user"; got != want {
		t.Fatalf("unexpected rest route URL: got %q want %q", got, want)
	}
}
