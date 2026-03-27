# Extension Testing Strategy

This document explains how to think about extension verification today, what is
already enforced by the platform, what is missing for external extension
authors, and what the recommended testing stack should be.

For the durable platform-side references behind this testing model, see:

- `../platform/docs/INSTANCE_AND_EXTENSION_LIFECYCLE.md`
- `../platform/docs/AGENT_CLI.md`
- `../platform/docs/EXTENSION_SECURITY_MODEL.md`

## Short Version

Today, the Move Big Rocks extension lifecycle already gives you some important
checks:

- `mbr extensions lint SOURCE_DIR`
- `mbr extensions lint SOURCE_DIR --write-contract`
- `mbr extensions verify SOURCE_DIR --workspace WORKSPACE_ID`
- `mbr extensions nav --instance --json`
- `mbr extensions widgets --instance --json`
- install from source or bundle
- `mbr extensions validate --id EXTENSION_ID`
- `mbr extensions activate --id EXTENSION_ID`
- `mbr extensions show --id EXTENSION_ID --json`
- `mbr extensions nav --workspace WORKSPACE_ID --json`
- `mbr extensions widgets --workspace WORKSPACE_ID --json`
- `mbr extensions monitor --id EXTENSION_ID`
- manifest validation in core
- runtime health and diagnostics for service-backed extensions

That is useful, but it is not yet a complete developer-facing test strategy.

The current system is strongest at:

- manifest contract validation
- route and endpoint validation
- health and runtime registration validation
- first-party regression tests inside the core repo

The current system is weakest at:

- reusable public harnesses that mirror first-party extension tests
- browser-level checks that prove the extension actually renders and appears in
  the admin shell
- explicit guidance for static admin pages that must keep working for instance
  admins without a live workspace context

## What Already Exists

### 1. SDK smoke script

The SDK already includes:

- [`scripts/verify-extension.sh`](./scripts/verify-extension.sh)

It runs:

- `mbr extensions validate --id EXTENSION_ID`
- `mbr extensions activate --id EXTENSION_ID`
- `mbr extensions show --id EXTENSION_ID --json`
- `mbr extensions monitor --id EXTENSION_ID --json`
- `mbr extensions nav --workspace WORKSPACE_ID --json` when `MBR_WORKSPACE_ID`
  is set
- `mbr extensions widgets --workspace WORKSPACE_ID --json` when
  `MBR_WORKSPACE_ID` is set
- `mbr extensions nav --instance --json`
- `mbr extensions widgets --instance --json`

That gives you a post-install smoke loop, but it is only one layer.

### 2. Core manifest and install validation

The platform already validates:

- required manifest fields
- runtime/storage combinations
- health endpoint requirements for service-backed packs
- asset existence for declared routes and endpoints
- admin navigation pointing at real `admin_page` endpoints
- dashboard widgets pointing at real `admin_page` endpoints
- event consumer and scheduled job contract rules
- route topology and path conflicts

Relevant code:

- `platform/internal/platform/domain/extension.go`
- `platform/internal/platform/services/extension_service.go`
- `platform/internal/platform/services/extension_runtime.go`

### 3. Runtime diagnostics and proof surfaces

The CLI can already expose useful runtime proof:

- `mbr extensions show --id EXTENSION_ID --json`
- `mbr extensions nav --workspace WORKSPACE_ID --json`
- `mbr extensions widgets --workspace WORKSPACE_ID --json`
- `mbr extensions monitor --id EXTENSION_ID --json`

Those surfaces already include:

- declared endpoints
- resolved admin navigation hrefs
- resolved dashboard widget hrefs
- seeded queue, form, and automation-rule inspection
- validation and health state
- runtime diagnostics for endpoints, consumers, and jobs

Relevant code:

- `platform/cmd/mbr/main.go`

### 4. Internal regression tests in the platform repo

The core repo already has good test coverage for the extension system itself:

- manifest validation tests
- admin navigation resolution tests
- route resolution tests
- install/activate/customize tests
- first-party pack install/activate tests

Relevant files:

- `platform/internal/platform/domain/extension_test.go`
- `platform/internal/platform/services/extension_service_test.go`
- `platform/internal/platform/services/extension_admin_navigation_test.go`
- `platform/internal/platform/services/extension_runtime_test.go`
- `platform/internal/platform/services/first_party_extension_packages_test.go`

## The Main Gap

The strongest helpers today live inside `platform/internal/...`.

That means first-party packs can benefit from rich regression tests, but a
third-party or customer extension repo built from this SDK cannot directly
reuse the same helpers.

In practice, external authors currently rely on:

- local Go tests they write themselves
- the install/validate/activate/monitor loop
- manual clicking in a sandbox

That is not enough if we want extension development to feel reliable and
contract-driven.

## Recommended Validation Stack

Use four layers.

### Layer 0: Extension-local unit tests

Every extension repo should own fast local tests for its own code:

- manifest JSON can unmarshal cleanly
- templates parse without runtime errors
- domain logic and stores work
- request handlers validate expected inputs

Examples for service-backed packs:

- parse embedded templates in a tiny `*_test.go`
- test stage seeding, vote dedupe, or slug generation
- test handler request/response behavior with `httptest`

This is the extension author’s fastest feedback loop.

### Layer 1: SDK contract verification

The SDK now has a public contract-verification layer.

1. `mbr extensions lint SOURCE_DIR`
   Runs offline checks on a source tree without installing it anywhere.

   It validates:

   - manifest schema and normalization rules
   - asset references
   - admin navigation and widget endpoint references
   - runtime/storage combinations
   - required health endpoints
   - command namespacing
   - event/job/consumer contract rules
   - `extension.contract.json` against the derived source contract

2. `mbr extensions lint SOURCE_DIR --write-contract`
   Refreshes `extension.contract.json` from the current manifest surface.

3. `mbr extensions verify SOURCE_DIR --workspace WORKSPACE_ID`
   Installs from source into a sandbox workspace, then runs a standard
   contract verification flow.

   It asserts:

   - the source passes `lint`
   - install succeeds
   - `validate` returns valid
   - `activate` succeeds
   - `monitor` returns healthy
   - `show` returns the expected declared endpoints, commands, skills, and
     assets
   - resolved admin navigation matches the contract
   - resolved dashboard widgets match the contract
   - instance-admin navigation still exposes the extension without a workspace
     selection when applicable
   - instance-admin dashboard widgets still expose the extension without a
     workspace selection when applicable
   - seeded resources exist and still match the manifest

4. Contract assertions file in the extension repo

   Example:

   - `extension.contract.json`

   This declares expected facts such as:

   - expected admin navigation hrefs
   - expected dashboard widgets
   - expected seeded queue slugs
   - expected seeded form slugs
   - expected public paths
   - expected runtime health endpoint

This is the missing bridge between the internal platform contract and the
external extension authoring lifecycle.

For workspace-scoped admin pages, the current platform rule is:

- service-backed admin pages receive the resolved install workspace even when an
  instance admin opens them without an active workspace
- instance-admin navigation hrefs for workspace-scoped installs should include a
  workspace hint so the intended install is unambiguous
- static asset admin pages that call workspace-bound APIs should preserve that
  workspace hint on their own API requests

### Layer 2: Sandbox smoke tests

Every serious extension should also have one sandbox smoke flow that exercises
the real installed pack.

Recommended checks:

- fetch the public page and confirm HTTP 200
- fetch the admin page with session auth and confirm HTTP 200
- hit the main extension API route if one exists
- verify one primary workflow end-to-end

Examples:

- sales pipeline: create a deal, move it, reload the board
- community feature requests: submit an idea, vote once, view it in admin

This layer proves that the extension works against a real running instance,
not just a manifest parser.

### Layer 3: Browser-level UI tests

For extensions with meaningful admin or public UI, add browser automation.

Recommended tool:

- Playwright

Use it for:

- confirming the extension appears in the admin navigation
- confirming clicking the menu entry lands on the expected page
- confirming major forms render and submit successfully
- confirming public pages render correctly

This is the only layer that truly proves “it appears in the menu and works in
the shell.”

## What Is Publicly Available Now

The platform now exposes the main proof surfaces external authors were missing:

- `mbr extensions show --id EXTENSION_ID --json`
  This now includes `resolvedAdminNavigation`, `resolvedDashboardWidgets`, and
  `seededResources`.
- `mbr extensions nav --workspace WORKSPACE_ID --json`
- `mbr extensions widgets --workspace WORKSPACE_ID --json`
- `mbr extensions nav --instance --json`
- `mbr extensions widgets --instance --json`

That means authors and agents can now prove:

- the extension resolved to the expected admin hrefs
- the extension contributed the expected dashboard widgets
- seeded queues, forms, and automation rules actually exist and still match the
  manifest

This closes one of the biggest gaps between first-party internal tests and
third-party extension authoring.

## What We Still Need To Expose Publicly

### 1. Public SDK test harness

The helper pattern used by:

- `platform/internal/testutil/refext/refext.go`

should exist in a public form.

Recommended shape:

- `github.com/movebigrocks/extension-sdk/testing/sdktest`

It should know how to:

- load an extension source tree
- install it into a sandbox
- activate it
- run standard assertions

### 2. Public behavior harnesses

The CLI and contract file now cover the structural proof layer.

What still needs a reusable public harness is the behavior layer:

- fetch public and admin pages over HTTP
- exercise the primary workflow
- run browser automation for menu visibility and form submission

## How API Changes Should Work

If we change the extension API or manifest contract, do it with explicit
versioning and canary packs.

### 1. Version the contract

Version at least these separately:

- manifest `schemaVersion`
- runtime contract version
- SDK verifier version

### 2. Update the SDK harness first

When the platform changes:

1. update manifest/runtime validation in core
2. update the SDK verifier
3. update reference extension fixtures
4. update first-party packs
5. only then tell external authors to move

That keeps the developer path coherent.

### 3. Keep first-party packs as canaries

The first-party extensions repo should be the compatibility canary.

Every platform change that affects extension contracts should run:

- manifest validation
- install/activate
- resolved navigation assertions
- route resolution assertions
- targeted workflow smoke tests

against all first-party packs.

If the first-party set fails, the contract change is not ready.

### 4. Support at least one older contract during migration

If a breaking change is unavoidable:

- add a new schema version
- keep the previous version supported for a migration window
- teach the SDK verifier to explain exactly what must change

That is much safer than silent breakage.

## Practical Recommendation

The recommended workflow is now:

1. Run `mbr extensions lint . --json` while editing.
2. If the declared extension surface changed intentionally, run
   `mbr extensions lint . --write-contract --json` and review the diff.
3. Run `mbr extensions verify . --workspace ws_preview --json`.
4. Exercise the main runtime workflow in the sandbox.
5. Add browser automation when the extension has meaningful UI.
6. Keep first-party and custom packs on the same contract loop.

That gives us one story for everyone:

- extension authors
- first-party pack maintainers
- customer private extension repos
- future platform contract changes

## Current Bottom Line

The SDK now gives you a real contract-first loop.

The platform already gives you strong manifest and runtime validation.

The main remaining gap is the reusable behavior harness on top:

- public HTTP smoke helpers
- public browser automation helpers
- public Go test fixtures instead of `platform/internal/...`
