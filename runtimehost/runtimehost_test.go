package runtimehost

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/movebigrocks/extension-sdk/runtimeproto"
)

func TestNewClientFromRequestUsesForwardedOrigin(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://unix/extensions/enterprise-access", nil)
	req.Header.Set(runtimeproto.HeaderHostToken, "tok_123")
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "api.movebigrocks.test")

	client, err := NewClientFromRequest(req)
	if err != nil {
		t.Fatalf("NewClientFromRequest returned error: %v", err)
	}
	if got := client.BaseURL; got != "https://api.movebigrocks.test" {
		t.Fatalf("expected forwarded base URL, got %q", got)
	}
	if got := client.Token; got != "tok_123" {
		t.Fatalf("expected host token, got %q", got)
	}
}

func TestIssueIdentitySession(t *testing.T) {
	var authHeader string
	var payload IdentitySessionRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(IdentitySessionResponse{
			UserID:           "usr_123",
			UserEmail:        payload.Email,
			UserName:         payload.Name,
			InstanceRole:     payload.InstanceRole,
			SessionToken:     "sess_123",
			SessionCreatedAt: time.Unix(10, 0).UTC(),
			SessionExpiresAt: time.Unix(20, 0).UTC(),
			CookieName:       "mbr_session",
			CookiePath:       "/",
			CookieSecure:     true,
			Message:          "existing user authenticated",
		})
	}))
	defer server.Close()

	client := &Client{
		BaseURL:    server.URL,
		Token:      "tok_123",
		HTTPClient: server.Client(),
	}
	resp, err := client.IssueIdentitySession(context.Background(), IdentitySessionRequest{
		Email:                "sso@example.com",
		Name:                 "SSO User",
		InstanceRole:         "operator",
		AllowJITProvisioning: true,
	})
	if err != nil {
		t.Fatalf("IssueIdentitySession returned error: %v", err)
	}
	if authHeader != "Bearer tok_123" {
		t.Fatalf("expected bearer auth header, got %q", authHeader)
	}
	if payload.Email != "sso@example.com" {
		t.Fatalf("expected request email to round-trip, got %q", payload.Email)
	}
	if resp.SessionToken != "sess_123" {
		t.Fatalf("expected session token, got %q", resp.SessionToken)
	}
}
