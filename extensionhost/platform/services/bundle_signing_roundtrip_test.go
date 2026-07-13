package platformservices

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"testing"

	platformdomain "github.com/movebigrocks/extension-sdk/extensionhost/platform/domain"
)

// TestSignedBundleRoundTripPublic proves the signer's canonicalization and the
// verifier's are the same one: a public bundle signed with the exported
// CanonicalSignedBundlePayload and BundleLicenseClaim must verify with
// VerifyBundle. This is the invariant the sign-bundle tool depends on, and it
// is what broke when the tool carried its own omitempty copy of the license
// claim: the license serialized differently on each side, so the signature
// never matched.
func TestSignedBundleRoundTripPublic(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	manifestRaw := json.RawMessage(`{"publisher":"DemandOps","slug":"ats","version":"0.8.35","kind":"product"}`)
	license := BundleLicenseClaim{
		Publisher: "DemandOps",
		Slug:      "ats",
		Version:   "0.8.35",
	}

	payload, err := CanonicalSignedBundlePayload(manifestRaw, nil, nil, license)
	if err != nil {
		t.Fatalf("canonical payload: %v", err)
	}
	signature := ed25519.Sign(priv, payload)

	bundle, err := json.Marshal(map[string]any{
		"manifest": manifestRaw,
		"trust": map[string]any{
			"keyID":     "demandops-public-1",
			"algorithm": "ed25519",
			"signature": base64.StdEncoding.EncodeToString(signature),
			"license":   license,
		},
	})
	if err != nil {
		t.Fatalf("marshal bundle: %v", err)
	}

	verifier, err := NewExtensionBundleTrustVerifier("mbr-prod-001", true, map[string]map[string]string{
		"DemandOps": {"demandops-public-1": base64.StdEncoding.EncodeToString(pub)},
	})
	if err != nil {
		t.Fatalf("build verifier: %v", err)
	}

	manifest := platformdomain.ExtensionManifest{Publisher: "DemandOps", Slug: "ats", Version: "0.8.35"}
	if err := verifier.VerifyBundle(context.Background(), manifest, "", bundle); err != nil {
		t.Fatalf("public bundle signed with the shared canonicalization must verify: %v", err)
	}
}

// TestSignedBundleRoundTripRejectsTamper guards that the round-trip is a real
// signature check, not a no-op: mutating the signed manifest must fail
// verification.
func TestSignedBundleRoundTripRejectsTamper(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	signedManifest := json.RawMessage(`{"publisher":"DemandOps","slug":"ats","version":"0.8.35","kind":"product"}`)
	license := BundleLicenseClaim{Publisher: "DemandOps", Slug: "ats", Version: "0.8.35"}
	payload, err := CanonicalSignedBundlePayload(signedManifest, nil, nil, license)
	if err != nil {
		t.Fatalf("canonical payload: %v", err)
	}
	signature := ed25519.Sign(priv, payload)

	tamperedManifest := json.RawMessage(`{"publisher":"DemandOps","slug":"ats","version":"0.8.35","kind":"privileged"}`)
	bundle, err := json.Marshal(map[string]any{
		"manifest": tamperedManifest,
		"trust": map[string]any{
			"keyID":     "demandops-public-1",
			"algorithm": "ed25519",
			"signature": base64.StdEncoding.EncodeToString(signature),
			"license":   license,
		},
	})
	if err != nil {
		t.Fatalf("marshal bundle: %v", err)
	}

	verifier, err := NewExtensionBundleTrustVerifier("mbr-prod-001", true, map[string]map[string]string{
		"DemandOps": {"demandops-public-1": base64.StdEncoding.EncodeToString(pub)},
	})
	if err != nil {
		t.Fatalf("build verifier: %v", err)
	}

	manifest := platformdomain.ExtensionManifest{Publisher: "DemandOps", Slug: "ats", Version: "0.8.35"}
	if err := verifier.VerifyBundle(context.Background(), manifest, "", bundle); err == nil {
		t.Fatal("a bundle whose manifest was changed after signing must fail verification")
	}
}
