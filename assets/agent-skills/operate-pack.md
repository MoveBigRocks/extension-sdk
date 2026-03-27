# Operate This Pack

Use this skill when an operator asks an agent to work with this extension.

## Default workflow

1. Inspect the manifest, `extension.contract.json`, and current config.
2. Run `mbr extensions lint . --json`.
3. Install or upgrade the pack in a sandbox workspace first, or use `mbr extensions verify . --workspace WORKSPACE_ID --json`.
4. Run:
   - `mbr extensions verify . --workspace WORKSPACE_ID --json`
   - or, for an already-installed pack:
   - `mbr extensions validate --id EXTENSION_ID`
   - `mbr extensions show --id EXTENSION_ID --json`
   - `mbr extensions nav --workspace WORKSPACE_ID --json`
   - `mbr extensions widgets --workspace WORKSPACE_ID --json`
   - `mbr extensions monitor --id EXTENSION_ID`
5. Exercise the main admin page and any public route.
6. Only then activate or keep it active in production.

## Do not do these blindly

- Do not activate a changed pack without reviewing the threat model.
- Do not install into production first when a sandbox is available.
- Do not assume this pack is safe for privileged auth or connector behavior unless the manifest and review explicitly say so.
