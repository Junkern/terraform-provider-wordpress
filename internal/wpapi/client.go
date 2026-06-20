package wpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
)

const (
	defaultContext  = "edit"
	defaultPerPage  = 100
	pageCollection  = "pages"
	postCollection  = "posts"
	userCollection  = "users"
	jsonContentType = "application/json"
)

// Client is a lightweight WordPress REST API client for page resources.
type Client struct {
	BaseURL    *url.URL
	Username   string
	Password   string
	HTTPClient *http.Client
}

// Page represents the WordPress page schema returned by the REST API.
type Page struct {
	ID            int64          `json:"id"`
	Date          string         `json:"date"`
	DateGMT       string         `json:"date_gmt"`
	GUID          RenderedField  `json:"guid"`
	Link          string         `json:"link"`
	Modified      string         `json:"modified"`
	ModifiedGMT   string         `json:"modified_gmt"`
	Slug          string         `json:"slug"`
	Status        string         `json:"status"`
	Type          string         `json:"type"`
	Password      string         `json:"password"`
	Parent        int64          `json:"parent"`
	Title         RenderedField  `json:"title"`
	Content       ContentField   `json:"content"`
	Author        int64          `json:"author"`
	Excerpt       ProtectedField `json:"excerpt"`
	FeaturedMedia int64          `json:"featured_media"`
	CommentStatus string         `json:"comment_status"`
	PingStatus    string         `json:"ping_status"`
	MenuOrder     int64          `json:"menu_order"`
	Template      string         `json:"template"`
	Meta          map[string]any `json:"meta,omitempty"`
}

// PageInput is used for create and update requests.
type PageInput struct {
	Date          *string `json:"date,omitempty"`
	DateGMT       *string `json:"date_gmt,omitempty"`
	Slug          *string `json:"slug,omitempty"`
	Status        *string `json:"status,omitempty"`
	Type          *string `json:"type,omitempty"`
	Password      *string `json:"password,omitempty"`
	Parent        *int64  `json:"parent,omitempty"`
	Title         *string `json:"title,omitempty"`
	Content       *string `json:"content,omitempty"`
	Author        *int64  `json:"author,omitempty"`
	Excerpt       *string `json:"excerpt,omitempty"`
	FeaturedMedia *int64  `json:"featured_media,omitempty"`
	CommentStatus *string `json:"comment_status,omitempty"`
	PingStatus    *string `json:"ping_status,omitempty"`
	MenuOrder     *int64  `json:"menu_order,omitempty"`
	Template      *string `json:"template,omitempty"`
}

// Post represents the WordPress post schema returned by the REST API.
type Post struct {
	ID            int64          `json:"id"`
	Date          string         `json:"date"`
	DateGMT       string         `json:"date_gmt"`
	GUID          RenderedField  `json:"guid"`
	Link          string         `json:"link"`
	Modified      string         `json:"modified"`
	ModifiedGMT   string         `json:"modified_gmt"`
	Slug          string         `json:"slug"`
	Status        string         `json:"status"`
	Type          string         `json:"type"`
	Password      string         `json:"password"`
	Title         RenderedField  `json:"title"`
	Content       ContentField   `json:"content"`
	Author        int64          `json:"author"`
	Excerpt       ProtectedField `json:"excerpt"`
	FeaturedMedia int64          `json:"featured_media"`
	CommentStatus string         `json:"comment_status"`
	PingStatus    string         `json:"ping_status"`
	Format        string         `json:"format"`
	Sticky        bool           `json:"sticky"`
	Template      string         `json:"template"`
	Meta          map[string]any `json:"meta,omitempty"`
}

// PostInput is used for create and update requests.
type PostInput struct {
	Date          *string `json:"date,omitempty"`
	DateGMT       *string `json:"date_gmt,omitempty"`
	Slug          *string `json:"slug,omitempty"`
	Status        *string `json:"status,omitempty"`
	Type          *string `json:"type,omitempty"`
	Password      *string `json:"password,omitempty"`
	Title         *string `json:"title,omitempty"`
	Content       *string `json:"content,omitempty"`
	Author        *int64  `json:"author,omitempty"`
	Excerpt       *string `json:"excerpt,omitempty"`
	FeaturedMedia *int64  `json:"featured_media,omitempty"`
	CommentStatus *string `json:"comment_status,omitempty"`
	PingStatus    *string `json:"ping_status,omitempty"`
	Format        *string `json:"format,omitempty"`
	Sticky        *bool   `json:"sticky,omitempty"`
	Template      *string `json:"template,omitempty"`
}

// User represents the WordPress user schema returned by the REST API.
type User struct {
	ID                int64             `json:"id"`
	Username          string            `json:"username"`
	Name              string            `json:"name"`
	FirstName         string            `json:"first_name"`
	LastName          string            `json:"last_name"`
	Email             string            `json:"email"`
	URL               string            `json:"url"`
	Description       string            `json:"description"`
	Link              string            `json:"link"`
	Locale            string            `json:"locale"`
	Nickname          string            `json:"nickname"`
	Slug              string            `json:"slug"`
	RegisteredDate    string            `json:"registered_date"`
	Roles             []string          `json:"roles"`
	Capabilities      map[string]bool   `json:"capabilities"`
	ExtraCapabilities map[string]bool   `json:"extra_capabilities"`
	AvatarURLs        map[string]string `json:"avatar_urls"`
	Meta              map[string]any    `json:"meta,omitempty"`
}

// UserInput is used for create and update requests.
type UserInput struct {
	Username    *string        `json:"username,omitempty"`
	Name        *string        `json:"name,omitempty"`
	FirstName   *string        `json:"first_name,omitempty"`
	LastName    *string        `json:"last_name,omitempty"`
	Email       *string        `json:"email,omitempty"`
	URL         *string        `json:"url,omitempty"`
	Description *string        `json:"description,omitempty"`
	Locale      *string        `json:"locale,omitempty"`
	Nickname    *string        `json:"nickname,omitempty"`
	Slug        *string        `json:"slug,omitempty"`
	Roles       []string       `json:"roles,omitempty"`
	Password    *string        `json:"password,omitempty"`
	Meta        map[string]any `json:"meta,omitempty"`
}

// RenderedField models a WordPress field with rendered and raw representations.
type RenderedField struct {
	Rendered string `json:"rendered"`
	Raw      string `json:"raw,omitempty"`
}

// ContentField models page content.
type ContentField struct {
	Rendered  string `json:"rendered"`
	Protected bool   `json:"protected,omitempty"`
	Raw       string `json:"raw,omitempty"`
}

// ProtectedField models excerpt content.
type ProtectedField struct {
	Rendered  string `json:"rendered"`
	Protected bool   `json:"protected,omitempty"`
	Raw       string `json:"raw,omitempty"`
}

// New creates a client for the given WordPress REST API base URL.
func New(rawBaseURL, username, password string) (*Client, error) {
	if strings.TrimSpace(rawBaseURL) == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	parsed, err := url.Parse(rawBaseURL)
	if err != nil {
		return nil, err
	}

	parsed.Path = strings.TrimSuffix(parsed.Path, "/")

	return &Client{
		BaseURL:  parsed,
		Username: username,
		Password: password,
	}, nil
}

// ListPages returns the collection of pages using the edit context.
func (c *Client) ListPages(ctx context.Context) ([]Page, error) {
	var pages []Page
	query := url.Values{}
	query.Set("context", defaultContext)
	query.Set("per_page", fmt.Sprintf("%d", defaultPerPage))

	if err := c.doJSON(ctx, http.MethodGet, c.requestURL(pageCollection, query), nil, &pages); err != nil {
		return nil, err
	}

	return pages, nil
}

// GetPage returns a single page by ID.
func (c *Client) GetPage(ctx context.Context, id int64) (*Page, error) {
	var page Page
	query := url.Values{}
	query.Set("context", defaultContext)

	if err := c.doJSON(ctx, http.MethodGet, c.requestURL(path.Join(pageCollection, fmt.Sprintf("%d", id)), query), nil, &page); err != nil {
		return nil, err
	}

	return &page, nil
}

// CreatePage creates a new page.
func (c *Client) CreatePage(ctx context.Context, input PageInput) (*Page, error) {
	var page Page
	if err := c.doJSON(ctx, http.MethodPost, c.requestURL(pageCollection, nil), input, &page); err != nil {
		return nil, err
	}

	return &page, nil
}

// UpdatePage updates an existing page.
func (c *Client) UpdatePage(ctx context.Context, id int64, input PageInput) (*Page, error) {
	var page Page
	if err := c.doJSON(ctx, http.MethodPost, c.requestURL(path.Join(pageCollection, fmt.Sprintf("%d", id)), nil), input, &page); err != nil {
		return nil, err
	}

	return &page, nil
}

// DeletePage deletes a page permanently.
func (c *Client) DeletePage(ctx context.Context, id int64) error {
	query := url.Values{}
	query.Set("force", "true")
	return c.doJSON(ctx, http.MethodDelete, c.requestURL(path.Join(pageCollection, fmt.Sprintf("%d", id)), query), nil, nil)
}

// ListPosts returns the collection of posts using the edit context.
func (c *Client) ListPosts(ctx context.Context) ([]Post, error) {
	var posts []Post
	query := url.Values{}
	query.Set("context", defaultContext)
	query.Set("per_page", fmt.Sprintf("%d", defaultPerPage))

	if err := c.doJSON(ctx, http.MethodGet, c.requestURL(postCollection, query), nil, &posts); err != nil {
		return nil, err
	}

	return posts, nil
}

// GetPost returns a single post by ID.
func (c *Client) GetPost(ctx context.Context, id int64) (*Post, error) {
	var post Post
	query := url.Values{}
	query.Set("context", defaultContext)

	if err := c.doJSON(ctx, http.MethodGet, c.requestURL(path.Join(postCollection, fmt.Sprintf("%d", id)), query), nil, &post); err != nil {
		return nil, err
	}

	return &post, nil
}

// CreatePost creates a new post.
func (c *Client) CreatePost(ctx context.Context, input PostInput) (*Post, error) {
	var post Post
	if err := c.doJSON(ctx, http.MethodPost, c.requestURL(postCollection, nil), input, &post); err != nil {
		return nil, err
	}

	return &post, nil
}

// UpdatePost updates an existing post.
func (c *Client) UpdatePost(ctx context.Context, id int64, input PostInput) (*Post, error) {
	var post Post
	if err := c.doJSON(ctx, http.MethodPost, c.requestURL(path.Join(postCollection, fmt.Sprintf("%d", id)), nil), input, &post); err != nil {
		return nil, err
	}

	return &post, nil
}

// DeletePost deletes a post permanently.
func (c *Client) DeletePost(ctx context.Context, id int64) error {
	query := url.Values{}
	query.Set("force", "true")
	return c.doJSON(ctx, http.MethodDelete, c.requestURL(path.Join(postCollection, fmt.Sprintf("%d", id)), query), nil, nil)
}

// ListUsers returns the collection of users using the edit context.
func (c *Client) ListUsers(ctx context.Context) ([]User, error) {
	var users []User
	query := url.Values{}
	query.Set("context", defaultContext)
	query.Set("per_page", fmt.Sprintf("%d", defaultPerPage))

	if err := c.doJSON(ctx, http.MethodGet, c.requestURL(userCollection, query), nil, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// GetUser returns a single user by ID.
func (c *Client) GetUser(ctx context.Context, id int64) (*User, error) {
	var user User
	query := url.Values{}
	query.Set("context", defaultContext)

	if err := c.doJSON(ctx, http.MethodGet, c.requestURL(path.Join(userCollection, fmt.Sprintf("%d", id)), query), nil, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// CreateUser creates a new user.
func (c *Client) CreateUser(ctx context.Context, input UserInput) (*User, error) {
	var user User
	if err := c.doJSON(ctx, http.MethodPost, c.requestURL(userCollection, nil), input, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateUser updates an existing user.
func (c *Client) UpdateUser(ctx context.Context, id int64, input UserInput) (*User, error) {
	var user User
	if err := c.doJSON(ctx, http.MethodPost, c.requestURL(path.Join(userCollection, fmt.Sprintf("%d", id)), nil), input, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// DeleteUser deletes a user and reassigns their posts.
func (c *Client) DeleteUser(ctx context.Context, id int64, reassign int64) error {
	query := url.Values{}
	query.Set("force", "true")
	query.Set("reassign", fmt.Sprintf("%d", reassign))
	return c.doJSON(ctx, http.MethodDelete, c.requestURL(path.Join(userCollection, fmt.Sprintf("%d", id)), query), nil, nil)
}

func (c *Client) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}

	return http.DefaultClient
}

func (c *Client) doJSON(ctx context.Context, method, rawURL string, body any, responseBody any) error {
	var requestBody io.Reader
	if body != nil {
		var buffer bytes.Buffer
		if err := json.NewEncoder(&buffer).Encode(body); err != nil {
			return err
		}
		requestBody = &buffer
	}

	req, err := http.NewRequestWithContext(ctx, method, rawURL, requestBody)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", jsonContentType)
	if body != nil {
		req.Header.Set("Content-Type", jsonContentType)
	}
	if strings.TrimSpace(c.Username) != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		trimmed := strings.TrimSpace(string(responseBytes))
		if trimmed != "" {
			return fmt.Errorf("wordpress %s %s returned %s: %s", method, rawURL, resp.Status, trimmed)
		}
		return fmt.Errorf("wordpress %s %s returned %s", method, rawURL, resp.Status)
	}

	if responseBody == nil || len(responseBytes) == 0 {
		return nil
	}

	if err := json.Unmarshal(responseBytes, responseBody); err != nil {
		return err
	}

	return nil
}

func (c *Client) requestURL(endpoint string, query url.Values) string {
	clone := *c.BaseURL
	clone.Path = path.Join(strings.TrimSuffix(clone.Path, "/"), endpoint)
	if len(query) > 0 {
		clone.RawQuery = query.Encode()
	}
	return clone.String()
}
