# Book Studio — Product specification

## Goal
An open-source, responsive editorial workspace to create, write, illustrate, edit, typeset, and prepare books for digital or print publication.

## Principles
- Real behavior only; no silent mocks or fake success states.
- Clear integration states when a provider is not configured.
- Author controls manuscript changes and publication decisions.
- Mobile, tablet, desktop and accessibility support.

## Primary journeys
1. Public site → sign in/up → onboarding.
2. Create project → select genre, audience, language, format and goals.
3. Plan outline → write chapters → save revisions.
4. Chat with editorial assistant → review/apply suggestions with undo.
5. Generate images/covers → select assets.
6. Configure layout → validate export requirements → export files.
7. Connect/disconnect Google Drive and manage synchronization.

## Persistence
Neon PostgreSQL is the primary relational source of truth. Google Drive stores user-authorized manuscripts and media. MongoDB is an optional secondary adapter for self-hosted VPS deployments, not the primary database.

## Authentication
Development-only UI access may be enabled by explicit environment flag outside production. Production requires Firebase Authentication and server-side authorization.

## AI
Use provider adapters and validate model capabilities (text, vision, image generation, tool calling, context length). Do not assume all NVIDIA models support every task.

## Quality gates
Type check, lint, unit/integration/E2E tests, project save/reload tests, provider failure tests, and export validation.