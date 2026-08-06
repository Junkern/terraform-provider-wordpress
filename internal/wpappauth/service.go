package wpappauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

const defaultApplicationName = "terraform-provider-wordpress"

var noncePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?s)var\s+wpApiSettings\s*=\s*\{"root":"([^"]+)","nonce":"([^"]+)"`),
	regexp.MustCompile(`name="_wpnonce"\s+value="([^"]+)"`),
	regexp.MustCompile(`name='_wpnonce'\s+value='([^']+)'`),
}

var themeInstallNoncePatterns = []*regexp.Regexp{
	regexp.MustCompile(`"_ajax_nonce":"([^"]+)"`),
	regexp.MustCompile(`"ajax_nonce":"([^"]+)"`),
	regexp.MustCompile(`"_ajax_nonce"\s*:\s*"([^"]+)"`),
	regexp.MustCompile(`"ajax_nonce"\s*:\s*"([^"]+)"`),
	regexp.MustCompile(`_ajax_nonce=([a-zA-Z0-9]+)`),
	regexp.MustCompile(`data-(?:ajax-nonce|wpnonce)="([^"]+)"`),
	regexp.MustCompile(`data-(?:ajax-nonce|wpnonce)='([^']+)'`),
	regexp.MustCompile(`name="_ajax_nonce"\s+value="([^"]+)"`),
	regexp.MustCompile(`name='_ajax_nonce'\s+value='([^']+)'`),
}

var themeActivationHrefPattern = regexp.MustCompile(`(?is)href\s*=\s*["']([^"']*themes\.php\?[^"']*)["']`)

// Service encapsulates the WordPress login and application password flow.
type Service struct {
	BaseURL         string
	Username        string
	Password        string
	ApplicationName string
	HTTPClient      *http.Client

	mu           sync.Mutex
	authedClient *http.Client
	siteURL      *url.URL
}

// Result is the response returned by WordPress when creating an application password.
type Result struct {
	UUID     string `json:"uuid"`
	AppID    string `json:"app_id"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// InstallTheme logs in with normal credentials and installs a theme via wp-admin/admin-ajax.php.
func (s *Service) InstallTheme(ctx context.Context, slug string) error {
	if strings.TrimSpace(slug) == "" {
		return errors.New("theme slug is required")
	}

	client, siteURL, err := s.ensureAuthenticatedSession(ctx)
	if err != nil {
		return err
	}

	nonce, err := s.fetchThemeInstallNonce(ctx, client, siteURL)
	if err != nil {
		return err
	}

	return s.installThemeAJAX(ctx, client, siteURL, slug, nonce)
}

// DeleteTheme logs in with normal credentials and deletes an installed theme via wp-admin/admin-ajax.php.
func (s *Service) DeleteTheme(ctx context.Context, slug string) error {
	if strings.TrimSpace(slug) == "" {
		return errors.New("theme slug is required")
	}

	client, siteURL, err := s.ensureAuthenticatedSession(ctx)
	if err != nil {
		return err
	}

	nonce, err := s.fetchThemeDeleteNonce(ctx, client, siteURL, slug)
	if err != nil {
		return err
	}

	return s.deleteThemeAJAX(ctx, client, siteURL, slug, nonce)
}

// ActivateTheme activates an installed theme through wp-admin.
func (s *Service) ActivateTheme(ctx context.Context, slug string) error {
	if strings.TrimSpace(slug) == "" {
		return errors.New("theme slug is required")
	}

	client, siteURL, err := s.ensureAuthenticatedSession(ctx)
	if err != nil {
		return err
	}

	activationURL, err := s.fetchThemeActivationURL(ctx, client, siteURL, slug)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, activationURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Referer", joinPath(siteURL, "wp-admin/themes.php"))
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("theme activation request returned %s", resp.Status)
	}

	return nil
}

func (s *Service) fetchThemeActivationURL(ctx context.Context, client *http.Client, siteURL *url.URL, slug string) (string, error) {
	themesURL := joinPath(siteURL, "wp-admin/themes.php")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, themesURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Referer", themesURL)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("themes page returned %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	for _, match := range themeActivationHrefPattern.FindAllStringSubmatch(html.UnescapeString(string(body)), -1) {
		if len(match) != 2 {
			continue
		}
		parsed, err := url.Parse(match[1])
		if err != nil || !strings.HasSuffix(parsed.Path, "/wp-admin/themes.php") {
			continue
		}
		query := parsed.Query()
		if query.Get("action") == "activate" && query.Get("stylesheet") == slug && query.Get("_wpnonce") != "" {
			return joinPath(siteURL, "wp-admin/themes.php") + "?" + query.Encode(), nil
		}
	}

	return "", fmt.Errorf("could not find activation link for theme %q", slug)
}

// CreateApplicationPassword logs in with the normal WordPress password and creates an application password.
func (s *Service) CreateApplicationPassword(ctx context.Context) (*Result, error) {
	if strings.TrimSpace(s.BaseURL) == "" {
		return nil, errors.New("base URL is required")
	}
	if strings.TrimSpace(s.Username) == "" {
		return nil, errors.New("username is required")
	}
	if strings.TrimSpace(s.Password) == "" {
		return nil, errors.New("password is required")
	}

	applicationName := strings.TrimSpace(s.ApplicationName)
	if applicationName == "" {
		applicationName = defaultApplicationName
	}

	client, siteURL, err := s.ensureAuthenticatedSession(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.loadLoginPage(ctx, client, siteURL); err != nil {
		return nil, err
	}
	if err := s.login(ctx, client, siteURL); err != nil {
		return nil, err
	}
	if err := s.ensureJSONRestAPI(ctx, client, siteURL); err != nil {
		return nil, err
	}

	state, err := s.fetchProfileState(ctx, client, siteURL)
	if err != nil {
		return nil, err
	}

	return s.createApplicationPassword(ctx, client, siteURL, state, applicationName)
}

func (s *Service) ensureAuthenticatedSession(ctx context.Context) (*http.Client, *url.URL, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.authedClient != nil && s.siteURL != nil {
		return s.authedClient, s.siteURL, nil
	}

	if err := s.validateCredentials(); err != nil {
		return nil, nil, err
	}

	client, err := s.client()
	if err != nil {
		return nil, nil, err
	}

	siteURL, err := siteBaseURL(s.BaseURL)
	if err != nil {
		return nil, nil, err
	}

	if err := s.loadLoginPage(ctx, client, siteURL); err != nil {
		return nil, nil, err
	}
	if err := s.login(ctx, client, siteURL); err != nil {
		return nil, nil, err
	}

	s.authedClient = client
	s.siteURL = siteURL

	return s.authedClient, s.siteURL, nil
}

func (s *Service) validateCredentials() error {
	if strings.TrimSpace(s.BaseURL) == "" {
		return errors.New("base URL is required")
	}
	if strings.TrimSpace(s.Username) == "" {
		return errors.New("username is required")
	}
	if strings.TrimSpace(s.Password) == "" {
		return errors.New("password is required")
	}

	return nil
}

func (s *Service) client() (*http.Client, error) {
	if s.HTTPClient != nil {
		if s.HTTPClient.Jar == nil {
			jar, err := cookiejar.New(nil)
			if err != nil {
				return nil, err
			}
			s.HTTPClient.Jar = jar
		}
		return s.HTTPClient, nil
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	return &http.Client{Jar: jar}, nil
}

func (s *Service) loadLoginPage(ctx context.Context, client *http.Client, siteURL *url.URL) error {
	loginURL := joinPath(siteURL, "wp-login.php?loggedout=true&wp_lang=en_US")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, loginURL, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("login page returned %s", resp.Status)
	}

	return nil
}

func (s *Service) login(ctx context.Context, client *http.Client, siteURL *url.URL) error {
	loginURL := joinPath(siteURL, "wp-login.php")
	form := url.Values{}
	form.Set("log", s.Username)
	form.Set("pwd", s.Password)
	form.Set("wp-submit", "Log In")
	form.Set("redirect_to", joinPath(siteURL, "wp-admin/"))
	form.Set("testcookie", "1")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", joinPath(siteURL, "wp-login.php?loggedout=true&wp_lang=en_US"))
	req.Header.Set("Origin", siteURL.String())
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("login request returned %s", resp.Status)
	}

	return nil
}

func (s *Service) ensureJSONRestAPI(ctx context.Context, client *http.Client, siteURL *url.URL) error {
	if ok, err := s.restAPIIsValid(ctx, client, siteURL); err != nil {
		return err
	} else if ok {
		return nil
	}

	if err := s.updatePermalinks(ctx, client, siteURL); err != nil {
		return err
	}

	ok, err := s.restAPIIsValid(ctx, client, siteURL)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("WordPress REST API still does not return valid JSON after updating permalinks")
	}

	return nil
}

func (s *Service) restAPIIsValid(ctx context.Context, client *http.Client, siteURL *url.URL) (bool, error) {
	restURL := joinPath(siteURL, "wp-json")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, restURL, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return false, nil
	}

	var decoded any
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return false, nil
	}

	return true, nil
}

func (s *Service) updatePermalinks(ctx context.Context, client *http.Client, siteURL *url.URL) error {
	pageURL := joinPath(siteURL, "wp-admin/options-permalink.php")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Referer", pageURL)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("permalink page returned %s", resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	body := html.UnescapeString(string(bodyBytes))
	nonce, err := extractHiddenField(body, "_wpnonce")
	if err != nil {
		return err
	}
	referer, err := extractHiddenField(body, "_wp_http_referer")
	if err != nil {
		return err
	}

	permalinkStructure := "/%year%/%monthnum%/%day%/%postname%/"
	form := url.Values{}
	form.Set("_wpnonce", nonce)
	form.Set("_wp_http_referer", referer)
	form.Set("selection", permalinkStructure)
	form.Set("permalink_structure", permalinkStructure)
	form.Set("category_base", "")
	form.Set("tag_base", "")
	form.Set("submit", "Save Changes")

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, pageURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Referer", pageURL)
	request.Header.Set("Origin", siteURL.String())
	request.Header.Set("User-Agent", "Mozilla/5.0")

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode < http.StatusOK || response.StatusCode >= 400 {
		return fmt.Errorf("permalink update returned %s", response.Status)
	}

	return nil
}

type profileState struct {
	RESTRoot string
	Nonce    string
	UserID   int
}

func (s *Service) fetchProfileState(ctx context.Context, client *http.Client, siteURL *url.URL) (*profileState, error) {
	profileURL := joinPath(siteURL, "wp-admin/profile.php")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, profileURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", joinPath(siteURL, "wp-admin/"))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("profile page returned %s", resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	body := html.UnescapeString(string(bodyBytes))
	state := &profileState{}
	for _, pattern := range noncePatterns {
		matches := pattern.FindStringSubmatch(body)
		if len(matches) == 3 {
			state.RESTRoot = matches[1]
			state.Nonce = matches[2]
			break
		}
		if len(matches) == 2 && state.Nonce == "" {
			state.Nonce = matches[1]
		}
	}

	userID, err := parseUserID(body)
	if err != nil {
		return nil, err
	}
	state.UserID = userID
	if state.RESTRoot == "" {
		state.RESTRoot = joinPath(siteURL, "index.php?rest_route=/")
	}
	if state.Nonce == "" {
		return nil, errors.New("could not find WordPress REST nonce on profile page")
	}

	return state, nil
}

func (s *Service) createApplicationPassword(ctx context.Context, client *http.Client, siteURL *url.URL, state *profileState, applicationName string) (*Result, error) {
	passwordURL := restRouteURL(state.RESTRoot, fmt.Sprintf("wp/v2/users/%d/application-passwords?_locale=user", state.UserID))
	form := url.Values{}
	form.Set("name", applicationName)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, passwordURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Referer", joinPath(siteURL, "wp-admin/profile.php?wp_http_referer=%2Fwp-admin%2Fusers.php"))
	req.Header.Set("X-WP-Nonce", state.Nonce)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Origin", siteURL.String())
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("application password request returned %s", resp.Status)
	}

	var result Result
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if strings.TrimSpace(result.Password) == "" {
		return nil, errors.New("application password response did not include a password")
	}

	return &result, nil
}

func (s *Service) fetchThemeInstallNonce(ctx context.Context, client *http.Client, siteURL *url.URL) (string, error) {
	themeInstallURL := joinPath(siteURL, "wp-admin/theme-install.php?browse=popular")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, themeInstallURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Referer", joinPath(siteURL, "wp-admin/"))
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("theme install page returned %s", resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	body := string(bodyBytes)

	// WordPress may render the login page instead of the theme installer when auth is missing.
	if resp.Request != nil && strings.Contains(resp.Request.URL.Path, "wp-login.php") {
		return "", errors.New("theme install request was redirected to wp-login.php; authenticated session may be missing or expired")
	}

	if strings.Contains(body, "not allowed to install themes") || strings.Contains(body, "you are not allowed to install themes") {
		return "", errors.New("authenticated user is not allowed to install themes")
	}

	nonce := themeNonceFromHTML(string(bodyBytes))
	if nonce != "" {
		return nonce, nil
	}

	return "", fmt.Errorf("could not find theme install ajax nonce on theme-install page (%s)", themeNonceDebugInfo(body))
}

func (s *Service) installThemeAJAX(ctx context.Context, client *http.Client, siteURL *url.URL, slug, nonce string) error {
	ajaxURL := joinPath(siteURL, "wp-admin/admin-ajax.php")
	form := url.Values{}
	form.Set("slug", slug)
	form.Set("action", "install-theme")
	form.Set("_ajax_nonce", nonce)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ajaxURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Referer", joinPath(siteURL, "wp-admin/theme-install.php?browse=popular"))
	req.Header.Set("Origin", siteURL.String())
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	body := strings.TrimSpace(string(bodyBytes))

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("install-theme ajax request returned %s: %s", resp.Status, body)
	}
	if body == "0" {
		return errors.New("install-theme ajax request was rejected")
	}

	var result struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(bodyBytes, &result); err == nil && !result.Success {
		return fmt.Errorf("install-theme ajax request failed: %s", body)
	}

	return nil
}

func (s *Service) fetchThemeDeleteNonce(ctx context.Context, client *http.Client, siteURL *url.URL, slug string) (string, error) {
	themeURL := joinPath(siteURL, "wp-admin/themes.php?theme="+url.QueryEscape(slug))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, themeURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Referer", joinPath(siteURL, "wp-admin/themes.php"))
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("theme detail page returned %s", resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	nonce := themeNonceFromHTML(string(bodyBytes))
	if nonce != "" {
		return nonce, nil
	}

	return "", errors.New("could not find theme delete ajax nonce on themes page")
}

func themeNonceFromHTML(body string) string {
	decoded := html.UnescapeString(body)
	for _, pattern := range themeInstallNoncePatterns {
		matches := pattern.FindStringSubmatch(decoded)
		if len(matches) == 2 && matches[1] != "" {
			return matches[1]
		}
	}

	return ""
}

func themeNonceDebugInfo(body string) string {
	decoded := html.UnescapeString(body)
	hints := []string{}

	if strings.Contains(decoded, "_wpUpdatesSettings") {
		hints = append(hints, "contains _wpUpdatesSettings")
	}
	if strings.Contains(decoded, "ajax_nonce") {
		hints = append(hints, "contains ajax_nonce")
	}
	if strings.Contains(decoded, "_ajax_nonce") {
		hints = append(hints, "contains _ajax_nonce")
	}
	if strings.Contains(decoded, "wpnonce") {
		hints = append(hints, "contains wpnonce")
	}

	window := 120
	re := regexp.MustCompile(`(?is).{0,120}nonce.{0,120}`)
	fragments := re.FindAllString(decoded, 2)
	for i, fragment := range fragments {
		trimmed := strings.Join(strings.Fields(fragment), " ")
		if len(trimmed) > window {
			trimmed = trimmed[:window] + "..."
		}
		hints = append(hints, fmt.Sprintf("fragment_%d=%q", i+1, trimmed))
	}

	if len(hints) == 0 {
		return "no nonce-like markers found"
	}

	return strings.Join(hints, "; ")
}

func (s *Service) deleteThemeAJAX(ctx context.Context, client *http.Client, siteURL *url.URL, slug, nonce string) error {
	ajaxURL := joinPath(siteURL, "wp-admin/admin-ajax.php")
	form := url.Values{}
	form.Set("slug", slug)
	form.Set("action", "delete-theme")
	form.Set("_ajax_nonce", nonce)
	form.Set("_fs_nonce", "")
	form.Set("username", "")
	form.Set("password", "")
	form.Set("connection_type", "")
	form.Set("public_key", "")
	form.Set("private_key", "")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ajaxURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Referer", joinPath(siteURL, "wp-admin/themes.php?theme="+url.QueryEscape(slug)))
	req.Header.Set("Origin", siteURL.String())
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	body := strings.TrimSpace(string(bodyBytes))

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("delete-theme ajax request returned %s: %s", resp.Status, body)
	}
	if body == "0" {
		return errors.New("delete-theme ajax request was rejected")
	}

	var result struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(bodyBytes, &result); err == nil && !result.Success {
		return fmt.Errorf("delete-theme ajax request failed: %s", body)
	}

	return nil
}

func parseUserID(body string) (int, error) {
	matches := regexp.MustCompile(`var\s+userId\s*=\s*([0-9]+);`).FindStringSubmatch(body)
	if len(matches) != 2 {
		return 0, errors.New("could not find WordPress user id on profile page")
	}

	userID, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func extractHiddenField(body string, fieldName string) (string, error) {
	pattern := regexp.MustCompile(fmt.Sprintf(`name="%s"\s+value="([^"]+)"`, regexp.QuoteMeta(fieldName)))
	matches := pattern.FindStringSubmatch(body)
	if len(matches) != 2 {
		return "", fmt.Errorf("could not find %s on WordPress settings page", fieldName)
	}

	return html.UnescapeString(matches[1]), nil
}

func restRouteURL(root string, route string) string {
	return strings.TrimRight(root, "/") + "/" + strings.TrimLeft(route, "/")
}

func siteBaseURL(raw string) (*url.URL, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}

	trimmed := strings.TrimSuffix(parsed.Path, "/")
	switch {
	case strings.HasSuffix(trimmed, "/wp-json/wp/v2"):
		parsed.Path = strings.TrimSuffix(trimmed, "/wp-json/wp/v2")
	case strings.HasSuffix(trimmed, "/wp-json"):
		parsed.Path = strings.TrimSuffix(trimmed, "/wp-json")
	default:
		parsed.Path = trimmed
	}

	if parsed.Path == "" {
		parsed.Path = "/"
	}
	if !strings.HasSuffix(parsed.Path, "/") {
		parsed.Path += "/"
	}
	parsed.RawQuery = ""
	parsed.Fragment = ""

	return parsed, nil
}

func joinPath(base *url.URL, rawReference string) string {
	reference, err := url.Parse(rawReference)
	if err != nil {
		return base.String()
	}
	return base.ResolveReference(reference).String()
}
