# Architecture

## Topology
- Next.js App Router: public site and application UI.
- Server-side API/service layer: authorization, project operations, provider adapters and job orchestration.
- Firebase Authentication: identity provider.
- Neon PostgreSQL: primary relational database.
- Google Drive API: user-authorized file storage.
- NVIDIA API: text and image provider adapter.
- Optional MongoDB adapter: secondary option for VPS deployments.

## Boundaries
UI components must not call external providers directly. Browser code must never receive provider secrets. Server-side services validate identity, ownership, inputs and permissions.

## Data ownership
Projects, chapters, assets, conversations and jobs must be scoped to a user/tenant. Server-side authorization is mandatory; hiding controls is not security.

## Synchronization
Use stable provider file IDs, revision metadata, idempotent jobs, bounded retries, and explicit conflict states. Distinguish editor state, database persistence and Drive synchronization.

## MongoDB
MongoDB is optional and secondary. Do not dual-write the same entity to PostgreSQL and MongoDB without an explicit consistency strategy.

## Security
Server-only secrets, encrypted OAuth tokens at rest, least-privilege Drive scopes, OAuth state validation, AI rate limits, audited sensitive operations, and production rejection of development bypass.