package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type bundleEnvelope struct {
	Manifest   json.RawMessage      `json:"manifest"`
	Assets     json.RawMessage      `json:"assets,omitempty"`
	Migrations json.RawMessage      `json:"migrations,omitempty"`
	Trust      *bundleTrustEnvelope `json:"trust,omitempty"`
}

type bundleTrustEnvelope struct {
	KeyID     string             `json:"keyID"`
	Algorithm string             `json:"algorithm,omitempty"`
	Signature string             `json:"signature"`
	License   bundleLicenseClaim `json:"license"`
}

type bundleLicenseClaim struct {
	InstanceID  string `json:"instanceID,omitempty"`
	Publisher   string `json:"publisher,omitempty"`
	Slug        string `json:"slug,omitempty"`
	Version     string `json:"version,omitempty"`
	TokenSHA256 string `json:"tokenSHA256,omitempty"`
}

type manifestSummary struct {
	Publisher string `json:"publisher"`
	Slug      string `json:"slug"`
	Version   string `json:"version"`
}

func main() {
	bundlePath := flag.String("bundle", "", "Unsigned bundle path")
	outPath := flag.String("out", "", "Signed bundle output path")
	keyID := flag.String("key-id", "", "Publisher key identifier")
	privateKeyBase64 := flag.String("private-key-base64", "", "Base64-encoded Ed25519 private key or seed")
	privateKeyEnv := flag.String("private-key-env", "", "Environment variable containing the base64-encoded Ed25519 private key or seed")
	instanceID := flag.String("instance-id", "", "Optional instance ID for an instance-bound signed bundle")
	licenseToken := flag.String("license-token", "", "Optional install credential for an instance-bound signed bundle")
	publisherKeyOut := flag.String("publisher-key-out", "", "Optional output path for the trusted publisher key JSON snippet")
	flag.Parse()

	if strings.TrimSpace(*bundlePath) == "" {
		exitf("missing --bundle")
	}
	if strings.TrimSpace(*outPath) == "" {
		exitf("missing --out")
	}
	if strings.TrimSpace(*keyID) == "" {
		exitf("missing --key-id")
	}
	if (strings.TrimSpace(*instanceID) == "") != (strings.TrimSpace(*licenseToken) == "") {
		exitf("pass both --instance-id and --license-token for an instance-bound signed bundle, or neither for a public signed bundle")
	}

	keyMaterial := strings.TrimSpace(*privateKeyBase64)
	if keyMaterial == "" && strings.TrimSpace(*privateKeyEnv) != "" {
		keyMaterial = strings.TrimSpace(os.Getenv(strings.TrimSpace(*privateKeyEnv)))
	}
	if keyMaterial == "" {
		exitf("missing signing key material")
	}
	privateKey, publicKey, err := decodePrivateKey(keyMaterial)
	if err != nil {
		exitf("decode private key: %v", err)
	}

	rawBundle, err := os.ReadFile(filepath.Clean(*bundlePath))
	if err != nil {
		exitf("read bundle: %v", err)
	}
	var envelope bundleEnvelope
	if err := json.Unmarshal(rawBundle, &envelope); err != nil {
		exitf("decode bundle: %v", err)
	}
	if len(envelope.Manifest) == 0 {
		exitf("bundle is missing manifest")
	}

	var manifest manifestSummary
	if err := json.Unmarshal(envelope.Manifest, &manifest); err != nil {
		exitf("decode manifest: %v", err)
	}
	if strings.TrimSpace(manifest.Publisher) == "" || strings.TrimSpace(manifest.Slug) == "" || strings.TrimSpace(manifest.Version) == "" {
		exitf("manifest must include publisher, slug, and version")
	}

	license := bundleLicenseClaim{
		Publisher: manifest.Publisher,
		Slug:      manifest.Slug,
		Version:   manifest.Version,
	}
	if strings.TrimSpace(*instanceID) != "" {
		license.InstanceID = strings.TrimSpace(*instanceID)
		license.TokenSHA256 = checksumSHA256Hex([]byte(strings.TrimSpace(*licenseToken)))
	}

	payload, err := canonicalSignedBundlePayload(envelope.Manifest, envelope.Assets, envelope.Migrations, license)
	if err != nil {
		exitf("build signed payload: %v", err)
	}
	signature := ed25519.Sign(privateKey, payload)
	envelope.Trust = &bundleTrustEnvelope{
		KeyID:     strings.TrimSpace(*keyID),
		Algorithm: "ed25519",
		Signature: base64.StdEncoding.EncodeToString(signature),
		License:   license,
	}

	signedBundle, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		exitf("encode signed bundle: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(filepath.Clean(*outPath)), 0o755); err != nil {
		exitf("create output directory: %v", err)
	}
	if err := os.WriteFile(filepath.Clean(*outPath), append(signedBundle, '\n'), 0o644); err != nil {
		exitf("write signed bundle: %v", err)
	}

	if strings.TrimSpace(*publisherKeyOut) != "" {
		keyJSON := map[string]map[string]string{
			manifest.Publisher: {
				strings.TrimSpace(*keyID): base64.StdEncoding.EncodeToString(publicKey),
			},
		}
		data, err := json.MarshalIndent(keyJSON, "", "  ")
		if err != nil {
			exitf("encode publisher key JSON: %v", err)
		}
		if err := os.MkdirAll(filepath.Dir(filepath.Clean(*publisherKeyOut)), 0o755); err != nil {
			exitf("create publisher key directory: %v", err)
		}
		if err := os.WriteFile(filepath.Clean(*publisherKeyOut), append(data, '\n'), 0o644); err != nil {
			exitf("write publisher key JSON: %v", err)
		}
	}
}

func decodePrivateKey(encoded string) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return nil, nil, err
	}
	switch len(keyBytes) {
	case ed25519.SeedSize:
		privateKey := ed25519.NewKeyFromSeed(keyBytes)
		return privateKey, privateKey.Public().(ed25519.PublicKey), nil
	case ed25519.PrivateKeySize:
		privateKey := ed25519.PrivateKey(keyBytes)
		return privateKey, privateKey.Public().(ed25519.PublicKey), nil
	default:
		return nil, nil, fmt.Errorf("expected %d-byte seed or %d-byte private key, got %d bytes", ed25519.SeedSize, ed25519.PrivateKeySize, len(keyBytes))
	}
}

func canonicalSignedBundlePayload(manifestRaw, assetsRaw, migrationsRaw json.RawMessage, license bundleLicenseClaim) ([]byte, error) {
	manifestValue, err := decodeBundleSection(manifestRaw, map[string]any{})
	if err != nil {
		return nil, fmt.Errorf("decode manifest: %w", err)
	}
	assetsValue, err := decodeBundleSection(assetsRaw, []any{})
	if err != nil {
		return nil, fmt.Errorf("decode assets: %w", err)
	}
	migrationsValue, err := decodeBundleSection(migrationsRaw, []any{})
	if err != nil {
		return nil, fmt.Errorf("decode migrations: %w", err)
	}
	return json.Marshal(map[string]any{
		"assets":     assetsValue,
		"license":    license,
		"manifest":   manifestValue,
		"migrations": migrationsValue,
	})
}

func decodeBundleSection(raw json.RawMessage, defaultValue any) (any, error) {
	if len(raw) == 0 {
		return defaultValue, nil
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	if value == nil {
		return defaultValue, nil
	}
	return value, nil
}

func checksumSHA256Hex(value []byte) string {
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:])
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
