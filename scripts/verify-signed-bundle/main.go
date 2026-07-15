// Command verify-signed-bundle checks that a signed bundle verifies against a
// set of trusted publisher keys, using the same VerifyBundle the core runs at
// install time. The SDK CI signs a sample bundle and then runs this, so a
// signer that produces bundles the core would reject fails the build instead of
// shipping. This is the end-to-end guard for the canonicalization drift that
// disabled extension trust in production.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/movebigrocks/extension-sdk/bundletrust"
)

func main() {
	bundlePath := flag.String("bundle", "", "Signed bundle path")
	trustedPath := flag.String("trusted-publishers", "", "Path to trusted publisher JSON (publisher -> keyID -> base64 public key)")
	instanceID := flag.String("instance-id", "", "Instance ID to verify against")
	licenseToken := flag.String("license-token", "", "Optional install credential for an instance-bound bundle")
	flag.Parse()

	if strings.TrimSpace(*bundlePath) == "" {
		exitf("missing --bundle")
	}
	if strings.TrimSpace(*trustedPath) == "" {
		exitf("missing --trusted-publishers")
	}
	if strings.TrimSpace(*instanceID) == "" {
		exitf("missing --instance-id")
	}

	bundle, err := os.ReadFile(filepath.Clean(*bundlePath))
	if err != nil {
		exitf("read bundle: %v", err)
	}

	var envelope struct {
		Manifest json.RawMessage `json:"manifest"`
	}
	if err := json.Unmarshal(bundle, &envelope); err != nil {
		exitf("decode bundle: %v", err)
	}
	var manifest bundletrust.ManifestIdentity
	if err := json.Unmarshal(envelope.Manifest, &manifest); err != nil {
		exitf("decode manifest: %v", err)
	}

	trustedRaw, err := os.ReadFile(filepath.Clean(*trustedPath))
	if err != nil {
		exitf("read trusted publishers: %v", err)
	}
	var trusted map[string]map[string]string
	if err := json.Unmarshal(trustedRaw, &trusted); err != nil {
		exitf("decode trusted publishers: %v", err)
	}

	verifier, err := bundletrust.NewVerifier(strings.TrimSpace(*instanceID), true, trusted)
	if err != nil {
		exitf("build verifier: %v", err)
	}
	if err := verifier.VerifyBundle(context.Background(), manifest, strings.TrimSpace(*licenseToken), bundle); err != nil {
		exitf("bundle failed verification: %v", err)
	}
	fmt.Printf("OK: %s/%s@%s verifies against %s\n", manifest.Publisher, manifest.Slug, manifest.Version, filepath.Base(*trustedPath))
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
