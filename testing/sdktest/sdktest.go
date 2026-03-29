package sdktest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

type CLI struct {
	Binary  string
	BaseURL string
	Token   string
	Env     []string
}

func (c CLI) binary() string {
	if strings.TrimSpace(c.Binary) != "" {
		return c.Binary
	}
	return "mbr"
}

func (c CLI) withDefaultFlags(args []string) []string {
	if strings.TrimSpace(c.BaseURL) != "" && !containsFlag(args, "--url") && !containsFlag(args, "--api-url") {
		args = append(args, "--url", c.BaseURL)
	}
	if strings.TrimSpace(c.Token) != "" && !containsFlag(args, "--token") {
		args = append(args, "--token", c.Token)
	}
	return args
}

func (c CLI) Run(ctx context.Context, args ...string) ([]byte, []byte, error) {
	cmd := exec.CommandContext(ctx, c.binary(), c.withDefaultFlags(append([]string(nil), args...))...)
	cmd.Env = append(os.Environ(), c.Env...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

func (c CLI) JSON(ctx context.Context, args ...string) (any, error) {
	stdout, stderr, err := c.Run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("run %q: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(stderr)))
	}
	var decoded any
	if err := json.Unmarshal(stdout, &decoded); err != nil {
		return nil, fmt.Errorf("decode json output: %w", err)
	}
	return decoded, nil
}

func (c CLI) JSONMap(ctx context.Context, args ...string) (map[string]any, error) {
	decoded, err := c.JSON(ctx, args...)
	if err != nil {
		return nil, err
	}
	value, ok := decoded.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected json object")
	}
	return value, nil
}

func (c CLI) JSONArray(ctx context.Context, args ...string) ([]map[string]any, error) {
	decoded, err := c.JSON(ctx, args...)
	if err != nil {
		return nil, err
	}
	items, ok := decoded.([]any)
	if !ok {
		return nil, fmt.Errorf("expected json array")
	}
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		entry, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("expected array of json objects")
		}
		result = append(result, entry)
	}
	return result, nil
}

func (c CLI) MustJSONMap(t testing.TB, ctx context.Context, args ...string) map[string]any {
	t.Helper()
	value, err := c.JSONMap(ctx, args...)
	if err != nil {
		t.Fatalf("json map command failed: %v", err)
	}
	return value
}

func (c CLI) MustJSONArray(t testing.TB, ctx context.Context, args ...string) []map[string]any {
	t.Helper()
	value, err := c.JSONArray(ctx, args...)
	if err != nil {
		t.Fatalf("json array command failed: %v", err)
	}
	return value
}

func ContainsKeyValue(items []map[string]any, key, want string) bool {
	for _, item := range items {
		value, ok := item[key]
		if !ok {
			continue
		}
		if text, ok := value.(string); ok && text == want {
			return true
		}
	}
	return false
}

func containsFlag(args []string, flag string) bool {
	for i := 0; i < len(args); i++ {
		if args[i] == flag {
			return true
		}
	}
	return false
}
