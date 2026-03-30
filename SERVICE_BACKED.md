# Service-Backed Upgrade Guide

Use this guide when a custom extension has outgrown the default bundle-first
shape and now genuinely needs backend runtime behavior.

## When To Use This Path

Stay bundle-first unless the extension needs one or more of these:

- custom server-side endpoint handlers
- event consumers
- scheduled jobs
- a health endpoint checked by the core runtime
- an owned `ext_*` PostgreSQL schema and migrations

Moving to `service_backed` is about runtime needs, not prestige. If the extension is
mostly static admin UI, public pages, workflow seeds, or bundled skills, keep
it bundle-first.

## Still Inside The Same Trust Slice

Going service-backed does **not** widen the allowed self-built trust model.

The supported generic self-built path remains:

- `scope: workspace`
- `risk: standard`
- `kind: product` or `kind: operational`

If the extension needs instance scope, privileged risk, identity behavior, or
connector behavior, stop and use a separately controlled path instead of
forcing it through the generic SDK flow.

## Minimum Shape

Keep the same repo shape and add the smallest runtime surface that proves the
need:

```text
.
├── manifest.json
├── extension.contract.json
├── assets/
├── runtime/
│   ├── main.go
│   ├── handlers.go
│   └── health_test.go
├── migrations/
│   └── 000001_init.up.sql
└── security/
```

Use that as a pattern, not a rigid framework requirement.

When you implement runtime code, prefer the public SDK helpers and standard Go
libraries. Do not reach into `platform/internal/...` from an external
extension repo.

## Manifest Changes

The typical upgrade from bundle-first to service-backed is:

1. change `runtimeClass` from `bundle` to `service_backed`
2. add a declared health endpoint
3. add declared runtime endpoints, consumers, jobs, or schema ownership only as
   needed
4. keep permissions and scope as small as possible

Example intent:

```json
{
  "runtimeClass": "service_backed",
  "storageClass": "owned_schema",
  "endpoints": [
    {
      "name": "my-extension-health",
      "class": "health",
      "mountPath": "/internal/extensions/my-extension/health",
      "methods": ["GET"]
    }
  ]
}
```

Do not cargo-cult this fragment. Match the actual fields and endpoint classes
supported by the current platform contract.

## Runtime Checklist

Your first service-backed pass should keep the runtime narrow:

- one health endpoint
- one real handler only if it is necessary
- one consumer or one scheduled job only if it is necessary
- one bounded schema with one initial migration only if it is necessary

That keeps the review surface understandable.

If you need a starting point for service-backed HTTP wiring, use the public SDK
runtime helpers such as `runtimehttp` rather than copying first-party helper
code into the extension repo.

## Validation Loop

Use this loop after converting the extension:

```bash
mbr extensions lint . --json
mbr extensions lint . --write-contract --json
mbr extensions verify . --workspace ws_preview --json
mbr extensions show --id EXTENSION_ID --json
mbr extensions monitor --id EXTENSION_ID --json
mbr extensions events list --workspace ws_preview --json
```

What you are proving now is broader than bundle-first:

- the manifest passes structural validation
- the health endpoint is declared and reachable
- runtime diagnostics expose handlers, consumers, or jobs
- schema or migration intent stays bounded and reviewable

## Tests You Should Add

At minimum, add:

- a runtime health test
- handler tests with `httptest` for each custom endpoint
- tests for consumer or scheduled-job behavior where relevant
- tests for migration collection if the extension owns SQL migrations
- one sandbox workflow smoke check through `mbr extensions verify`

If the extension has meaningful UI, add browser automation as well. See
[`examples/playwright/`](./examples/playwright/).

## Review Gates

Before activating a service-backed extension outside preview:

- rerun `security/threat-model.md`
- rerun `security/review-checklist.md`
- review declared endpoints, jobs, consumers, and schema ownership
- review external calls and secrets
- confirm rollback or deactivation steps are explicit

Service-backed extensions are still bounded extensions on the shared base. They do
not bypass shared auth, shared audit, shared routing, or the sanctioned core
action paths.
