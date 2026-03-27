# Move Big Rocks Extension SDK Template

This tree is the canonical Move Big Rocks extension-SDK layout for building extensions
on top of Move Big Rocks, the AI-native service operations platform.

It is licensed under BSL 1.1 with the same no-resale rule used across the
public Move Big Rocks code and extension surfaces. See `LICENSE`.

It is the canonical starting point for a custom extension repo that lives
outside the public Move Big Rocks core repo.

## What This Source Tree Is

This source tree is:

- the starting point for private custom-extension repos
- the starting point for free public extensions built on the same contract
- the authoring contract for builder workflows
- the place agents and humans should read before creating a custom extension repo

This source tree is not:

- a live extension repo for customer production state
- the whole public core repo
- the deployment control plane for a Move Big Rocks instance

Start here:

- [START_HERE.md](START_HERE.md)
- [TESTING.md](TESTING.md)

The default contract-first loop is now:

```bash
mbr extensions lint . --json
mbr extensions verify . --workspace ws_preview --json
mbr extensions nav --instance --json
mbr extensions widgets --instance --json
```

Treat instance-admin behavior as part of the contract, not a nice-to-have. If a
workspace-scoped extension exposes admin pages, an instance admin with no active
workspace should still see the pack in instance navigation and open a working
entrypoint.

If you intentionally change the declared extension surface, refresh the
contract file and review the diff:

```bash
mbr extensions lint . --write-contract --json
```

What this template gives you:

- one clear agent handoff file
- one testing and verification guide
- one machine-readable `extension.contract.json`
- one proof-oriented validation loop for resolved navigation, widgets, and seeded resources
- one explicit instance-admin/no-workspace validation expectation
- one valid bundle-first `manifest.json`
- one minimal admin page
- one minimal public page
- one bundled agent skill
- one threat-model prompt
- one review checklist
- one sandbox install script
- one sandbox upgrade script
- one activation and monitor script
- one bundle build script
- one signing-key generation script
- one bundle signing script
- one OCI publish script

This template is intentionally the simplest safe authoring path:

- workspace-scoped
- standard-risk
- bundle-first

## Distribution Model

This template is meant to support the current public Move Big Rocks extension
policy:

- build your extension privately by default
- package it as a signed bundle for repeatable delivery
- install it from source, a local bundle file, an HTTPS URL, or an OCI ref
- keep it private or give it away for free if you decide to publish it
- do not plan on selling Move Big Rocks extensions without separate written
  permission from Move Big Rocks BV

## Public Bundle Tooling

The SDK now includes the same basic tooling the first-party public bundle flow
uses:

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

For a public signed bundle, omit `--instance-id` and `--license-token` from the
signing step. For an instance-bound signed bundle, pass both.

The generated seed belongs in your publishing environment as
`MBR_EXTENSION_SIGNING_PRIVATE_KEY_B64`. The generated trusted-publisher JSON
belongs in the instance config as `EXTENSION_TRUSTED_PUBLISHERS_JSON`.

If you publish to GHCR, remember that the first package publication may still
need its visibility changed to `Public` in GitHub Packages before anonymous
pulls work as intended.

If you need service-backed behavior later, keep the same repo shape and add:

- service-backed endpoints
- health endpoint
- event consumers
- scheduled jobs
- owned-schema migrations

Do not start there unless you actually need it.
