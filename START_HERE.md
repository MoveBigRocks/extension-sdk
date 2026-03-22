# Start Here

This file is the one an agent should read first when helping a user build a custom Move Big Rocks extension.

## Goal

Build a safe extension that can be:

1. developed locally
2. installed into a sandbox workspace
3. validated
4. activated
5. monitored
6. upgraded or deactivated cleanly later

## Default Rule

Start with the simplest possible extension:

- `workspace` scope
- `standard` risk
- `bundle` runtime

Only move to a service-backed extension if the requirement truly needs:

- custom server-side handlers
- event consumers
- scheduled jobs
- owned Postgres schema and migrations

## Files That Matter

- `manifest.json`
- `assets/admin/dashboard.html`
- `assets/public/index.html`
- `assets/agent-skills/operate-pack.md`
- `security/threat-model.md`
- `security/review-checklist.md`
- `scripts/install-into-sandbox.sh`
- `scripts/upgrade-in-sandbox.sh`
- `scripts/verify-extension.sh`

## First Pass Workflow

1. Rename the extension in `manifest.json`:
   - `slug`
   - `name`
   - `publisher`
   - any example routes
2. Keep the runtime bundle-first unless the requirement clearly needs more.
3. Implement the first useful admin or public page.
4. Add one bundled agent skill that explains how to operate the pack.
5. Threat-model the extension before activation.

## Local Development Loop

Assume:

- you already have a local or staging Move Big Rocks instance
- you can log in with the CLI
- you have a sandbox workspace to test in

Recommended loop:

```bash
mbr auth login --url https://app.yourdomain.com
mbr workspaces list
mbr extensions install . --workspace ws_sandbox --license-token lic_sandbox
mbr extensions list --workspace ws_sandbox
mbr extensions validate EXTENSION_ID
mbr extensions activate EXTENSION_ID
mbr extensions monitor --id EXTENSION_ID
mbr extensions skills list --id EXTENSION_ID
```

The helper scripts wrap the same lifecycle for agents:

```bash
export MBR_URL=https://app.yourdomain.com
export MBR_WORKSPACE_ID=ws_sandbox
export MBR_LICENSE_TOKEN=lic_sandbox
./scripts/install-into-sandbox.sh .

export MBR_EXTENSION_ID=ext_installed_id
./scripts/verify-extension.sh
./scripts/upgrade-in-sandbox.sh .
```

If the extension changes:

```bash
mbr extensions upgrade . --id EXTENSION_ID --license-token lic_sandbox
mbr extensions monitor --id EXTENSION_ID
```

If something looks wrong:

```bash
mbr extensions deactivate EXTENSION_ID
```

## Review Rule

Do not activate a self-built extension outside a sandbox workspace until both of these are complete:

- `security/threat-model.md`
- `security/review-checklist.md`

## Packaging Rule

This template can be installed directly from the source directory during development.

For marketplace or repeatable delivery later, package it into a bundle and install the bundle by file, HTTPS URL, OCI ref, or marketplace alias.

## What "Done" Means

An extension is only done when:

- it installs cleanly
- validation passes
- activation succeeds
- monitor reports healthy
- the main workflow can be exercised without undocumented steps
- the threat model and review checklist are complete
