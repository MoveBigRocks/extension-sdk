# Operate This Pack

Use this skill when an operator asks an agent to work with this extension.

## Default workflow

1. Inspect the manifest and current config.
2. Install or upgrade the pack in a sandbox workspace first.
3. Run:
   - `mbr extensions validate EXTENSION_ID`
   - `mbr extensions monitor --id EXTENSION_ID`
4. Exercise the main admin page and any public route.
5. Only then activate or keep it active in production.

## Do not do these blindly

- Do not activate a changed pack without reviewing the threat model.
- Do not install into production first when a sandbox is available.
- Do not assume this pack is safe for privileged auth or connector behavior unless the manifest and review explicitly say so.
