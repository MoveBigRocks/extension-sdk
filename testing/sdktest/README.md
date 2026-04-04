# sdktest

`sdktest` is a small public helper package for writing extension-repo smoke
tests around the supported `mbr extensions ... --json` lifecycle.

It is intentionally narrow. The goal is to make the contract-first extension
loop easier to reuse from a custom extension repo without importing anything
from `github.com/movebigrocks/platform/...`.

## What It Helps With

- running `mbr` with a fixed base URL and optional token
- decoding JSON results from extension lifecycle commands
- checking instance or workspace navigation and widget surfaces
- writing lightweight smoke tests around `verify`, `show`, and `monitor`

## Example

```go
package extensiontest

import (
  "context"
  "testing"

  "github.com/movebigrocks/extension-sdk/testing/sdktest"
)

func TestExtensionLifecycle(t *testing.T) {
  cli := sdktest.CLI{BaseURL: "https://app.example.com"}
  ctx := context.Background()

  result := cli.MustJSONMap(t, ctx, "extensions", "verify", ".", "--workspace", "ws_preview", "--json")
  if result["valid"] == nil {
    t.Fatalf("expected verify response to include validation state")
  }
}
```

Keep this helper for contract-level smoke checks. Extension-specific workflow
behavior should still have its own tests in the extension repo.
