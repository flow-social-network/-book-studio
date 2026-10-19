# Book Studio

Open-source AI-powered book creation studio: writing, illustration, editorial review, typesetting, and publishing workflows.

## Stack
- Next.js App Router, React, TypeScript
- Tailwind CSS
- Firebase Authentication (integration phase)
- Neon PostgreSQL (primary relational database)
- Google Drive API (user-authorized file storage)
- NVIDIA API (provider adapter for text and image models)
- MongoDB (optional secondary adapter for self-hosted VPS deployments)

## Development
Requires Node.js 22+ and pnpm.

1. Copy .env.example to .env.local.
2. Run pnpm install.
3. Run pnpm dev.

The initial UI foundation may display explicitly labeled example projects. These are not persisted user data. Integrations must never report success unless the service confirms it.

## Authentication
Production must use real Firebase Authentication and server-side authorization. A temporary development-only UI bypass may be used only outside production, behind an explicit environment flag. It is not authentication and must never be enabled in production.

## Docs
See docs/PRODUCT_SPEC.md, docs/ARCHITECTURE.md, docs/ROUTES.md, and docs/DEVELOPMENT_STATUS.md.

License has not yet been selected.