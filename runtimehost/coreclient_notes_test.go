package runtimehost

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCaseReceiptContractsAndRejectedRedirect(t *testing.T) {
	var key, body string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == CoreCasesPath {
			var in CreateCaseInput
			_ = json.NewDecoder(r.Body).Decode(&in)
			key = in.IdempotencyKey
			_ = json.NewEncoder(w).Encode(HostCase{ID: "case-1"})
			return
		}
		if r.URL.Path == CoreCasesPath+"/case-1/notes" {
			var in AppendCaseNoteInput
			_ = json.NewDecoder(r.Body).Decode(&in)
			key = in.IdempotencyKey
			body = in.Body
			_ = json.NewEncoder(w).Encode(HostCaseNote{ID: "note-1", CaseID: "case-1"})
			return
		}
		http.Redirect(w, r, CoreCasesPath, http.StatusTemporaryRedirect)
	}))
	defer server.Close()
	c := &Client{BaseURL: server.URL, Token: "host-test"}
	_, err := c.CreateCase(context.Background(), CreateCaseInput{IdempotencyKey: "stable-case", Subject: "outage"})
	if err != nil || key != "stable-case" {
		t.Fatalf("create contract: %s %v", key, err)
	}
	note, err := c.AppendCaseNote(context.Background(), "case-1", AppendCaseNoteInput{IdempotencyKey: "stable-note", Body: "Evidence"})
	if err != nil || key != "stable-note" || body != "Evidence" || note.ID != "note-1" {
		t.Fatalf("note contract: %v", err)
	}
	_, _, err = c.GetCase(context.Background(), "redirect")
	if StatusCode(err) != http.StatusTemporaryRedirect {
		t.Fatalf("redirect must remain a rejection: %v", err)
	}
}
func TestHostResponseBound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(strings.Repeat("x", (2<<20)+1))) }))
	defer server.Close()
	c := &Client{BaseURL: server.URL, Token: "host-test"}
	_, _, err := c.GetCase(context.Background(), "case")
	if err == nil || !strings.Contains(err.Error(), "size limit") {
		t.Fatalf("oversized host response: %v", err)
	}
}
