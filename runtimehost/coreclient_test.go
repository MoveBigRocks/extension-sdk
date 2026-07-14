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

func TestClientCreateQueueAndUpdateCasePaths(t *testing.T) {
	var method, path string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			_ = json.NewEncoder(w).Encode(HostQueue{ID: "q1", Slug: "triage"})
		} else {
			_ = json.NewEncoder(w).Encode(HostCase{ID: "case-1"})
		}
	}))
	defer server.Close()
	c := &Client{BaseURL: server.URL, Token: "tok"}

	if _, err := c.CreateQueue(context.Background(), CreateQueueInput{Name: "Triage", Slug: "triage"}); err != nil {
		t.Fatalf("CreateQueue: %v", err)
	}
	if method != http.MethodPost || path != CoreQueuesPath {
		t.Fatalf("CreateQueue hit %s %s", method, path)
	}

	status := "resolved"
	if _, err := c.UpdateCase(context.Background(), "case-1", CaseUpdateInput{Status: &status}); err != nil {
		t.Fatalf("UpdateCase: %v", err)
	}
	if method != http.MethodPatch || path != CoreCasesPath+"/case-1" {
		t.Fatalf("UpdateCase hit %s %s", method, path)
	}
}
