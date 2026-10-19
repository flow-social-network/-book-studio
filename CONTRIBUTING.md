# Contributing to Book Studio

Thank you for considering contributing to **Book Studio**! We welcome code, documentation, design, accessibility improvements, testing, bug reports, thoughtful feature proposals, and help improving the editorial workflow.

Book Studio is an open-source editorial studio in active development. Contributions should help make the platform reliable, understandable, secure, accessible, and useful to authors and publishing teams.

- Project: https://github.com/flow-social-network/-book-studio
- Issues: https://github.com/flow-social-network/-book-studio/issues
- Pull requests: https://github.com/flow-social-network/-book-studio/pulls
- Donations and project support: [DONATIONS.md](DONATIONS.md)

## Before you start

1. Read the [README](README.md), [product specification](docs/PRODUCT_SPEC.md), [architecture](docs/ARCHITECTURE.md), and [development status](docs/DEVELOPMENT_STATUS.md).
2. Search existing issues and pull requests to avoid duplicate work.
3. For a substantial change, open an issue first to discuss the problem, intended behavior, and architectural impact.
4. Keep each pull request focused on one coherent change.

## Ways to contribute

- **Code:** application flows, editor behavior, accessibility, persistence, provider adapters, exports, and tests.
- **Documentation:** setup instructions, architecture records, API contracts, troubleshooting, and examples.
- **Quality assurance:** reproducible bug reports, regression tests, browser/device testing, and failure-path testing.
- **Design:** responsive layouts, accessible interaction patterns, and consistent design-system components.
- **Community:** help triage issues, clarify requirements, and review proposed changes.

## Development workflow

Use a supported Node.js version (22+ recommended) and pnpm.

```bash
git clone https://github.com/flow-social-network/-book-studio.git
cd -book-studio
pnpm install
cp .env.example .env.local
pnpm dev
```

Before opening a pull request, run the applicable checks:

```bash
pnpm typecheck
pnpm lint
pnpm build
```

Report which commands you ran and their actual results. Do not claim a check passed if it was not executed. If a command currently fails because the project is incomplete, include the exact error and context.

## Engineering standards

- Respect the existing architecture and document significant decisions.
- Use TypeScript types and clear contracts at module boundaries.
- Validate inputs at system boundaries and handle errors explicitly.
- Keep provider secrets and OAuth secrets on the server; never commit credentials.
- Never add real user manuscripts, private documents, access tokens, API keys, or personal data to the repository.
- Do not introduce fake success states or sample records that could be mistaken for persisted data.
- Distinguish local editor state, database persistence, file export, and Google Drive synchronization.
- Do not report an external publication as complete unless the destination service confirms it.
- Keep PostgreSQL as the primary persistence path. MongoDB is an optional secondary adapter and must not become a hidden mandatory dependency.
- Add or update tests for new behavior, error cases, and authorization boundaries.
- Keep UI changes responsive and keyboard-accessible.
- Update documentation and `.env.example` when configuration changes.

## Pull request checklist

- [ ] The problem and intended behavior are explained.
- [ ] The change is scoped and avoids unrelated refactors.
- [ ] Tests were added or updated where applicable.
- [ ] Type checking, linting, and build checks were run, and their real results are reported.
- [ ] No secrets, personal data, generated user content, or unnecessary binaries were committed.
- [ ] Error, loading, empty, and success states are truthful.
- [ ] Documentation and environment variable examples are updated.
- [ ] Any database migration or compatibility concern is described.
- [ ] Screenshots or a short recording are attached for meaningful UI changes.

## Commit and pull request guidance

Use descriptive commit messages, for example:

- `feat(editor): add chapter outline navigation`
- `fix(drive): handle expired authorization`
- `test(export): cover invalid manuscript metadata`
- `docs(architecture): clarify provider boundaries`

A pull request should describe:

1. The problem being solved.
2. The proposed approach and important trade-offs.
3. How the change was tested.
4. Any migration, environment, security, or deployment impact.
5. Any known limitation or follow-up work.

Be respectful and constructive in code review. We may ask for changes to preserve consistency, maintainability, security, or scope.

## License and contribution terms

Unless a file explicitly states otherwise, contributions intended for inclusion in Book Studio are submitted under the **GNU Affero General Public License v3.0 (AGPL-3.0-only)**, the project license. By submitting a contribution, you represent that you have the right to submit it under those terms and that you have identified any third-party material and its applicable license.

You retain copyright in your contributions; this statement does not transfer copyright ownership to the project maintainer. Do not submit code copied from sources whose license or terms are incompatible or unclear. Contributions from employers or other rights holders must be authorized.

No separate Contributor License Agreement (CLA) is currently defined. If that policy changes, the repository will document it before requiring contributors to agree to new terms.

## Security issues

Please do not publish exploitable vulnerabilities, secrets, or private user data in public issues. Until a dedicated private security reporting channel is configured, avoid posting sensitive details publicly and open a minimal issue asking the maintainers for a private contact route.

## Community and support

Open a GitHub issue for reproducible bugs and feature proposals. For project funding, see [DONATIONS.md](DONATIONS.md). Donations are voluntary, do not purchase influence over technical decisions, and do not guarantee a feature, service level, support response, or delivery date.

Thank you for helping make Book Studio better for authors and the open-source community.