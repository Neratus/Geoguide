package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"testing"
	"time"

	"github.com/Neratus/geoguide/internal/repository"
	"github.com/Neratus/geoguide/internal/repository/config"
)

const (
	TestUser     = "testuser"
	TestPassword = "TestPass123!"
)

type TestClient struct {
	baseURL string
	client  *http.Client
	t       *testing.T
	token   string
}

type testEnv struct {
	dbURL  string
	appURL string
	cfg    *config.Config
	repos  *repository.Repositories
}

func NewTestClient(t *testing.T, env *testEnv) *TestClient {
	t.Helper()
	jar, _ := cookiejar.New(nil)
	return &TestClient{
		baseURL: env.appURL,
		client: &http.Client{
			Jar:     jar,
			Timeout: 10 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		t: t,
	}
}

func (c *TestClient) Login(username, password string) error {
	c.t.Helper()
	body, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/auth/login", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("login request: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed: status=%d, body=%s", resp.StatusCode, string(raw))
	}
	var result struct {
		Token             string `json:"token"`
		RequiresTwoFactor bool   `json:"requiresTwoFactor"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return fmt.Errorf("decode login response: %w, body=%s", err, string(raw))
	}
	if result.RequiresTwoFactor {
		return fmt.Errorf("login requires 2FA, but test user should not have 2FA enabled")
	}
	if result.Token == "" {
		return fmt.Errorf("login returned empty token, body=%s", string(raw))
	}
	c.token = result.Token
	return nil
}

func (c *TestClient) Logout() error {
	c.t.Helper()
	if c.token == "" {
		return nil
	}
	resp, err := c.DoAPI("POST", "/api/v1/auth/logout", nil)
	if err != nil {
		return fmt.Errorf("logout request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("logout: status=%d, body=%s", resp.StatusCode, string(body))
	}
	c.token = ""
	return nil
}

func (c *TestClient) DoAPI(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}
	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	return c.client.Do(req)
}

func DecodeJSON(t *testing.T, resp *http.Response, target interface{}) {
	t.Helper()
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/json") && !strings.Contains(ct, "json") {
		t.Fatalf("expected JSON, got %q (status=%d): %s", ct, resp.StatusCode, string(raw))
	}
	if err := json.Unmarshal(raw, target); err != nil {
		t.Fatalf("decode JSON: %v (status=%d, body=%s)", err, resp.StatusCode, string(raw))
	}
}

func ReadBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(raw)
}

func WithUser(t *testing.T, env *testEnv, username, password string, fn func(c *TestClient)) {
	t.Helper()
	c := NewTestClient(t, env)
	if err := c.Login(username, password); err != nil {
		t.Fatalf("login as %s failed: %v", username, err)
	}
	t.Cleanup(func() {
		if err := c.Logout(); err != nil {
			t.Logf("logout as %s failed (non-fatal): %v", username, err)
		}
	})
	fn(c)
}
