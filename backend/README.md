# Book Studio Backend

Independent Go HTTP API. It owns authorization, domain validation, database persistence, AI provider calls, asset metadata, editorial reviews and EPUB generation. The web app communicates through this API; it must not access PostgreSQL directly.

## Implemented API surface

All /api/v1 endpoints require a valid Firebase ID token in the Authorization: Bearer token header.

- GET /api/v1/dashboard — real per-user counts and recent projects.
- GET and POST /api/v1/projects — list and create projects.
- GET, PATCH and DELETE /api/v1/projects/{id} — read, update and delete a project.
- GET and POST /api/v1/projects/{id}/chapters — list and create chapters.
- PATCH and DELETE /api/v1/projects/{id}/chapters/{chapterID} — update/delete chapters.
- POST /api/v1/projects/{id}/ai — real NVIDIA NIM request for writing assistance or editorial review.
- GET /api/v1/reviews — persisted review history.
- GET /api/v1/projects/{id}/export/epub — generate an EPUB from persisted content.
- GET and POST /api/v1/assets — list and upload real assets.
- GET and DELETE /api/v1/assets/{id} — delete asset.
- GET /api/v1/assets/{id}/download — download an owned asset.
- GET and PUT /api/v1/me/preferences — per-user preferences.
- GET /api/v1/integrations — configuration and operational status.
- GET /health and GET /ready — health and PostgreSQL readiness.

## Required configuration

1. Copy .env.example to .env and fill real server-side values.
2. Set DATABASE_URL to a Neon PostgreSQL connection string.
3. Set FIREBASE_PROJECT_ID and FIREBASE_CREDENTIALS_FILE to the Firebase project and service-account JSON path. Never commit the service-account file.
4. Set NVIDIA_API_KEY to use AI endpoints.
5. Set ASSET_STORAGE_DIR to a persistent volume for uploaded files. An ephemeral filesystem is not suitable for production.
6. Set CORS_ALLOWED_ORIGINS to exact frontend origins.
7. Run the API from this directory so the migration runner can read migrations/0001_initial.sql.

The service refuses to start when PostgreSQL is not configured or unavailable. Database schema migration is applied at startup. Protected requests fail closed when Firebase token verification fails.

## Run locally

    cp .env.example .env
    # Fill all required credentials in .env.
    mkdir -p /var/lib/book-studio/assets
    go mod tidy
    go test ./...
    go run ./cmd/api

## Limitations that must remain explicit

- Google Drive OAuth, token storage and file synchronization are not implemented yet.
- PDF export, print layout engine and direct publishing integrations are not implemented yet.
- AI requests require an actual NVIDIA key and supported model; no AI response is fabricated when the provider fails.
- Upload storage currently uses ASSET_STORAGE_DIR; use a persistent volume. Configure a managed object store before deploying to ephemeral/serverless compute.
- Firebase account login, registration and password recovery are client SDK flows; backend authorization verifies Firebase ID tokens.

Do not mark production-ready until CI passes, migrations are tested against a disposable PostgreSQL database, Firebase and NVIDIA credentials are configured in a secure environment, and end-to-end authorization tests pass.
