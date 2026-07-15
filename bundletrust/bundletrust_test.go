package bundletrust

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestVerifierAcceptsCanonicalPublicBundle(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	manifest := json.RawMessage(`{"publisher":"DemandOps","slug":"sample","version":"1.0.0"}`)
	assets := json.RawMessage(`[{"path":"index.html"}]`)
	license := LicenseClaim{Publisher: "DemandOps", Slug: "sample", Version: "1.0.0"}
	payload, err := CanonicalSignedBundlePayload(manifest, assets, nil, license)
	if err != nil {
		t.Fatalf("canonical payload: %v", err)
	}
	bundle, err := json.Marshal(map[string]any{
		"manifest": json.RawMessage(manifest),
		"assets":   json.RawMessage(assets),
		"trust": map[string]any{
			"keyID": "primary", "algorithm": "ed25519",
			"signature": base64.StdEncoding.EncodeToString(ed25519.Sign(privateKey, payload)),
			"license":   license,
		},
	})
	if err != nil {
		t.Fatalf("encode bundle: %v", err)
	}
	verifier, err := NewVerifier("instance-1", true, map[string]map[string]string{
		"DemandOps": {"primary": base64.StdEncoding.EncodeToString(publicKey)},
	})
	if err != nil {
		t.Fatalf("new verifier: %v", err)
	}
	if err := verifier.VerifyBundle(context.Background(), ManifestIdentity{Publisher: "DemandOps", Slug: "sample", Version: "1.0.0"}, "", bundle); err != nil {
		t.Fatalf("VerifyBundle: %v", err)
	}

	bundle[len(bundle)-2] ^= 1
	if err := verifier.VerifyBundle(context.Background(), ManifestIdentity{Publisher: "DemandOps", Slug: "sample", Version: "1.0.0"}, "", bundle); err == nil {
		t.Fatal("expected a tampered bundle to be rejected")
	}
}
