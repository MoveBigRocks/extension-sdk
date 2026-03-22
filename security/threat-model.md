# Extension Threat Model

Complete this before activating the extension outside a sandbox workspace.

## Extension identity

- Slug:
- Version:
- Publisher:
- Runtime class:
- Scope:
- Risk:

## Capabilities requested

- Permissions:
- Public endpoints:
- Admin endpoints:
- Agent skills:
- Commands:

## Data and storage

- Which Move Big Rocks primitives does this pack touch?
- Does it create forms, queues, or automation?
- Does it store secrets?
- Does it call external systems?
- If service-backed, what schema and migrations does it own?

## Failure and rollback

- What breaks if activation fails?
- What breaks if the extension is unhealthy?
- Can it be safely deactivated?
- Is there any destructive uninstall behavior?

## Security notes

- Authentication model:
- Input validation risks:
- Cross-workspace or cross-instance risks:
- External webhook or ingest abuse risks:
- Logging and data-exposure risks:
