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

func TestClientAttachmentAndArtifactPaths(t *testing.T) {
	var method, path string
	var gotContent []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		if r.Method == http.MethodPost && r.URL.Path == CoreAttachmentsPath {
			var in UploadAttachmentInput
			_ = json.NewDecoder(r.Body).Decode(&in)
			gotContent = in.Content
			_ = json.NewEncoder(w).Encode(HostAttachment{ID: "att-1", Filename: in.Filename})
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	c := &Client{BaseURL: server.URL, Token: "tok"}

	att, err := c.UploadAttachment(context.Background(), UploadAttachmentInput{Filename: "r.pdf", Content: []byte("bytes")})
	if err != nil {
		t.Fatalf("UploadAttachment: %v", err)
	}
	if att.ID != "att-1" || method != http.MethodPost || path != CoreAttachmentsPath {
		t.Fatalf("UploadAttachment hit %s %s -> %+v", method, path, att)
	}
	if string(gotContent) != "bytes" {
		t.Fatalf("content did not round-trip through base64: %q", gotContent)
	}

	if err := c.PublishArtifact(context.Background(), PublishArtifactInput{Surface: "website", RelativePath: "i.html", Content: []byte("h")}); err != nil {
		t.Fatalf("PublishArtifact: %v", err)
	}
	if method != http.MethodPost || path != CoreArtifactsPath {
		t.Fatalf("PublishArtifact hit %s %s", method, path)
	}
}

func TestClientIngestApplicationPath(t *testing.T) {
	var method, path string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		_ = json.NewEncoder(w).Encode(IngestApplicationResult{ContactID: "c1", CaseID: "k1"})
	}))
	defer server.Close()
	c := &Client{BaseURL: server.URL, Token: "tok"}
	out, err := c.IngestApplication(context.Background(), IngestApplicationInput{
		IdempotencyKey: "app-1",
		Contact:        CreateContactInput{Email: "a@b.com"},
		Case:           IngestCaseInput{Subject: "Application"},
	})
	if err != nil {
		t.Fatalf("IngestApplication: %v", err)
	}
	if out.ContactID != "c1" || out.CaseID != "k1" {
		t.Fatalf("unexpected result: %+v", out)
	}
	if method != http.MethodPost || path != CoreIngestApplicationPath {
		t.Fatalf("IngestApplication hit %s %s", method, path)
	}
}

func TestClientApplyCaseChangePath(t *testing.T) {
	var method, path string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		_ = json.NewEncoder(w).Encode(HostCase{ID: "case-1"})
	}))
	defer server.Close()
	c := &Client{BaseURL: server.URL, Token: "tok"}
	if _, err := c.ApplyCaseChange(context.Background(), "case-1", ApplyCaseChangeInput{IdempotencyKey: "k", Event: "stage"}); err != nil {
		t.Fatalf("ApplyCaseChange: %v", err)
	}
	if method != http.MethodPost || path != CoreCasesPath+"/case-1/apply-change" {
		t.Fatalf("ApplyCaseChange hit %s %s", method, path)
	}
}
