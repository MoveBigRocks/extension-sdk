# Playwright Smoke Example

This folder is a minimal example for extension repos that want one browser-level
smoke check on top of the contract-first CLI loop.

Use it when the extension has meaningful UI and you want to prove:

- it appears in the admin shell
- the menu entry opens the expected page
- the primary public or admin screen renders

## What This Example Assumes

- the extension is already installed in a preview workspace
- `mbr extensions verify . --workspace WORKSPACE_ID --json` already passes
- you can authenticate to the target Move Big Rocks instance in a browser

## Environment

Set:

- `MBR_BASE_URL` such as `https://app.example.com`
- `MBR_EXTENSION_PATH` such as `/extensions/sample-ops-extension`
- optionally `MBR_PUBLIC_PATH` such as `/sample-ops-extension`

## Run

```bash
cd examples/playwright
npm install
npx playwright test
```

The example is intentionally small. Treat it as a starter, not as a framework.
