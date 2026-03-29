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

## Prerequisites

Before using this SDK, make sure you already have:

- a running Move Big Rocks instance you control locally, in staging, or in a
  preview environment
- the `mbr` CLI installed and authenticated against that instance
- one preview or sandbox workspace where the extension can be installed safely

If you do not have those yet, stop here and use the deployment path first:

- [MoveBigRocks/platform/docs/CUSTOMER_INSTANCE_SETUP.md](https://github.com/MoveBigRocks/platform/blob/main/docs/CUSTOMER_INSTANCE_SETUP.md)
- [MoveBigRocks/platform/docs/INSTANCE_AND_EXTENSION_LIFECYCLE.md](https://github.com/MoveBigRocks/platform/blob/main/docs/INSTANCE_AND_EXTENSION_LIFECYCLE.md)
- [movebigrocks.com/docs/self-host](https://movebigrocks.com/docs/self-host)

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

## Supported Generic Runtime Slice

This SDK is the supported public starting point for self-built extensions that
stay inside the current generic runtime envelope:

- `scope: workspace`
- `risk: standard`
- `kind: product` or `kind: operational`

Do not use this path for currently restricted generic categories such as:

- `scope: instance`
- `risk: privileged`
- `kind: identity`
- `kind: connector`

You can still move from `bundle` to `service_backed` later when the extension
needs backend handlers, jobs, consumers, or an owned schema, but it should
still remain inside the supported trust slice above unless a first-party or
separately controlled path is being used.

## Files That Matter

- `manifest.json`
- `extension.contract.json`
- `assets/admin/dashboard.html`
- `assets/public/index.html`
- `assets/agent-skills/operate-pack.md`
- `TESTING.md`
- `SERVICE_BACKED.md`
- `security/threat-model.md`
- `security/review-checklist.md`
- `scripts/install-into-sandbox.sh`
- `scripts/upgrade-in-sandbox.sh`
- `scripts/verify-extension.sh`
- `scripts/build-bundle.go`
- `scripts/generate-signing-key.go`
- `scripts/sign-bundle.go`
- `scripts/publish-bundle-oci.sh`
- `examples/playwright/`

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
mbr extensions lint . --json
mbr auth login --url https://app.yourdomain.com
mbr workspaces list
mbr extensions verify . --workspace ws_sandbox --json
mbr extensions nav --instance --json
mbr extensions widgets --instance --json
mbr extensions skills list --id EXTENSION_ID
```

Then read `TESTING.md` and make sure the extension can prove the main workflow,
not just install and activate cleanly.

If the requirement genuinely needs backend runtime behavior, stop and read
`SERVICE_BACKED.md` before inventing your own runtime shape.

Important rule for workspace-scoped admin pages:

- do not assume the user will always have a live workspace session context
- an instance admin with no active workspace should still be able to discover
  the extension and open a working entrypoint
- if your admin UI is static-asset based and calls workspace-bound APIs, carry
  the `?workspace=...` hint through those API requests

The helper scripts wrap the same lifecycle for agents:

```bash
export MBR_URL=https://app.yourdomain.com
export MBR_WORKSPACE_ID=ws_sandbox
./scripts/install-into-sandbox.sh .

export MBR_EXTENSION_SOURCE_DIR=.
./scripts/verify-extension.sh

export MBR_EXTENSION_ID=ext_installed_id
./scripts/upgrade-in-sandbox.sh .
```

If your organization is using a controlled instance-bound bundle flow, you can
still export `MBR_LICENSE_TOKEN` before running the helper scripts.

If the extension changes:

```bash
mbr extensions lint . --write-contract --json
mbr extensions verify . --workspace ws_sandbox --json
```

If something looks wrong:

```bash
mbr extensions deactivate --id EXTENSION_ID
```

Public signed bundles and local source installs do not need an instance-bound
token. Keep `--license-token` for controlled instance-bound bundle flows.

## Review Rule

Do not activate a self-built extension outside a sandbox workspace until both of these are complete:

- `security/threat-model.md`
- `security/review-checklist.md`

## Packaging Rule

This template can be installed directly from the source directory during development.

For repeatable delivery later, package it into a signed bundle and install the
bundle by file, HTTPS URL, or OCI ref. Marketplace aliases are optional and
only matter when a private catalog is in use.

## Publication Rule

Under the current public Move Big Rocks license and distribution policy, custom
extensions can stay private or be given away for free. Selling Move Big Rocks
extensions requires separate written permission from Move Big Rocks BV.

The bundled publication tooling supports both public signed bundles and
instance-bound signed bundles:

```bash
go run ./scripts/generate-signing-key.go \
  --publisher DemandOps \
  --key-id demandops-public-1 \
  --seed-out secrets/demandops-public-1.seed.b64 \
  --trusted-publishers-out dist/demandops-public-1.publisher.json
go run ./scripts/build-bundle.go --source . --out dist/my-extension.bundle.json
go run ./scripts/sign-bundle.go \
  --bundle dist/my-extension.bundle.json \
  --out dist/my-extension.signed.bundle.json \
  --key-id demandops-public-1 \
  --private-key-env MBR_EXTENSION_SIGNING_PRIVATE_KEY_B64
./scripts/publish-bundle-oci.sh \
  --bundle dist/my-extension.signed.bundle.json \
  --image ghcr.io/movebigrocks/mbr-ext-my-extension \
  --tag v0.1.0
```

If you are publishing a free public bundle, you do not need an instance-bound
license claim. If you are publishing a controlled private bundle, pass
`--instance-id` and `--license-token` to `sign-bundle.go`.

Put the generated seed into your CI or publishing environment as
`MBR_EXTENSION_SIGNING_PRIVATE_KEY_B64`. Put the generated trusted publisher
JSON into the instance config as `EXTENSION_TRUSTED_PUBLISHERS_JSON`.

## What "Done" Means

An extension is only done when:

- it installs cleanly
- `mbr extensions lint .` passes
- validation passes
- activation succeeds
- monitor reports healthy
- `extension.contract.json` matches the real extension surface
- `mbr extensions nav --instance --json` and `mbr extensions widgets --instance --json`
  still show the extension when that makes sense
- the main workflow can be exercised without undocumented steps
- the threat model and review checklist are complete
