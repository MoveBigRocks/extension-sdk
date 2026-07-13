package runtimehost

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientCreateCaseRoundTrip(t *testing.T) {
	var gotAuth, gotPath, gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(HostCase{ID: "case-1", WorkspaceID: "ws-1", Subject: "Boom"})
	}))
	defer server.Close()

	c := &Client{BaseURL: server.URL, Token: "tok-123"}
	out, err := c.CreateCase(context.Background(), CreateCaseInput{Subject: "Boom", Priority: "high"})
	if err != nil {
		t.Fatalf("CreateCase: %v", err)
	}
	if out.ID != "case-1" {
		t.Fatalf("expected case-1, got %q", out.ID)
	}
	if gotAuth != "Bearer tok-123" {
		t.Fatalf("expected bearer auth, got %q", gotAuth)
	}
	if gotPath != CoreCasesPath {
		t.Fatalf("expected path %q, got %q", CoreCasesPath, gotPath)
	}
	var sent CreateCaseInput
	if err := json.Unmarshal([]byte(gotBody), &sent); err != nil {
		t.Fatalf("decode sent body: %v", err)
	}
	if sent.Subject != "Boom" || sent.Priority != "high" {
		t.Fatalf("request body not sent faithfully: %+v", sent)
	}
}

func TestClientGetCaseNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Status: "failed", Message: "core entity not found"})
	}))
	defer server.Close()

	c := &Client{BaseURL: server.URL, Token: "tok-123"}
	got, ok, err := c.GetCase(context.Background(), "missing")
	if err != nil {
		t.Fatalf("GetCase should map 404 to (nil,false,nil), got err %v", err)
	}
	if ok || got != nil {
		t.Fatalf("expected not-found, got ok=%v case=%+v", ok, got)
	}
}
