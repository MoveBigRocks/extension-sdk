package sdktest

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestWithDefaultFlagsAddsBaseURLAndToken(t *testing.T) {
	cli := CLI{BaseURL: "https://app.example.com", Token: "hat_example"}
	args := cli.withDefaultFlags([]string{"extensions", "show", "--id", "ext_123", "--json"})

	requireContainsSequence(t, args, []string{"--url", "https://app.example.com"})
	requireContainsSequence(t, args, []string{"--token", "hat_example"})
}

func TestWithDefaultFlagsDoesNotDuplicateExplicitFlags(t *testing.T) {
	cli := CLI{BaseURL: "https://app.example.com", Token: "hat_example"}
	args := cli.withDefaultFlags([]string{"extensions", "show", "--url", "https://other.example.com", "--token", "hat_other", "--json"})

	if countFlag(args, "--url") != 1 {
		t.Fatalf("expected one --url flag, got %d", countFlag(args, "--url"))
	}
	if countFlag(args, "--token") != 1 {
		t.Fatalf("expected one --token flag, got %d", countFlag(args, "--token"))
	}
}

func TestJSONMapDecodesCommandOutput(t *testing.T) {
	binary := writeStubCommand(t, `printf '{"status":"ok","kind":"verify"}'`)
	cli := CLI{Binary: binary}

	value, err := cli.JSONMap(context.Background(), "extensions", "verify", ".", "--json")
	if err != nil {
		t.Fatalf("json map: %v", err)
	}
	if value["status"] != "ok" {
		t.Fatalf("status = %v, want ok", value["status"])
	}
}

func TestJSONArrayDecodesCommandOutput(t *testing.T) {
	binary := writeStubCommand(t, `printf '[{"title":"Sample Ops Extension"}]'`)
	cli := CLI{Binary: binary}

	value, err := cli.JSONArray(context.Background(), "extensions", "nav", "--instance", "--json")
	if err != nil {
		t.Fatalf("json array: %v", err)
	}
	if !ContainsKeyValue(value, "title", "Sample Ops Extension") {
		t.Fatalf("expected nav output to contain Sample Ops Extension")
	}
}

func writeStubCommand(t *testing.T, body string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell stub helper is unix-only")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "mbr-stub.sh")
	script := "#!/usr/bin/env bash\nset -euo pipefail\n" + body + "\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write stub command: %v", err)
	}
	return path
}

func countFlag(args []string, flag string) int {
	count := 0
	for _, arg := range args {
		if arg == flag {
			count++
		}
	}
	return count
}

func requireContainsSequence(t *testing.T, args []string, want []string) {
	t.Helper()
	for i := 0; i <= len(args)-len(want); i++ {
		match := true
		for j := range want {
			if args[i+j] != want[j] {
				match = false
				break
			}
		}
		if match {
			return
		}
	}
	t.Fatalf("args %v do not contain sequence %v", args, want)
}
