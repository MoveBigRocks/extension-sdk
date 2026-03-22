# Move Big Rocks Extension SDK Template

This tree is the canonical Move Big Rocks extension-SDK layout for building extensions
on top of Move Big Rocks, the AI-native service operations platform.

It is the canonical starting point for a private custom extension repo inside
the public Move Big Rocks core repo.

## What This Source Tree Is

This source tree is:

- the starting point for private custom-extension repos
- the authoring contract for builder workflows
- the place agents and humans should read before creating a custom extension repo

This source tree is not:

- a live extension repo for customer production state
- the whole public core repo
- the deployment control plane for a Move Big Rocks instance

Start here:

- [START_HERE.md](START_HERE.md)

What this template gives you:

- one clear agent handoff file
- one valid bundle-first `manifest.json`
- one minimal admin page
- one minimal public page
- one bundled agent skill
- one threat-model prompt
- one review checklist
- one sandbox install script
- one sandbox upgrade script
- one activation and monitor script

This template is intentionally the simplest safe authoring path:

- workspace-scoped
- standard-risk
- bundle-first

If you need service-backed behavior later, keep the same repo shape and add:

- service-backed endpoints
- health endpoint
- event consumers
- scheduled jobs
- owned-schema migrations

Do not start there unless you actually need it.
