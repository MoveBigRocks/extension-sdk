package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	publisher := flag.String("publisher", "", "Publisher name for trusted publisher JSON")
	keyID := flag.String("key-id", "", "Publisher key identifier")
	seedOut := flag.String("seed-out", "", "Optional path for the base64-encoded Ed25519 seed")
	trustedPublishersOut := flag.String("trusted-publishers-out", "", "Optional path for trusted publisher JSON output")
	flag.Parse()

	if strings.TrimSpace(*publisher) == "" {
		exitf("missing --publisher")
	}
	if strings.TrimSpace(*keyID) == "" {
		exitf("missing --key-id")
	}

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		exitf("generate key: %v", err)
	}

	seedBase64 := base64.StdEncoding.EncodeToString(privateKey.Seed())
	publicKeyBase64 := base64.StdEncoding.EncodeToString(publicKey)
	trustedPublishers := map[string]map[string]string{
		strings.TrimSpace(*publisher): {
			strings.TrimSpace(*keyID): publicKeyBase64,
		},
	}
	trustedPublishersJSON, err := json.MarshalIndent(trustedPublishers, "", "  ")
	if err != nil {
		exitf("encode trusted publisher JSON: %v", err)
	}

	if strings.TrimSpace(*seedOut) != "" {
		if err := writeTextFile(*seedOut, seedBase64+"\n"); err != nil {
			exitf("write seed: %v", err)
		}
	}
	if strings.TrimSpace(*trustedPublishersOut) != "" {
		if err := writeTextFile(*trustedPublishersOut, string(trustedPublishersJSON)+"\n"); err != nil {
			exitf("write trusted publisher JSON: %v", err)
		}
	}

	fmt.Printf("MBR_EXTENSION_SIGNING_PRIVATE_KEY_B64=%s\n\n", seedBase64)
	fmt.Printf("EXTENSION_TRUSTED_PUBLISHERS_JSON=%s\n", trustedPublishersJSON)
}

func writeTextFile(path string, content string) error {
	cleanPath := filepath.Clean(path)
	if err := os.MkdirAll(filepath.Dir(cleanPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(cleanPath, []byte(content), 0o600)
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
