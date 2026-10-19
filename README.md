# Book Studio

**Book Studio** é um estúdio editorial com IA para planear, escrever, ilustrar, rever, diagramar, exportar e preparar livros para publicação.

## Identidade visual e metadados
- Logótipo vetorial: `public/brand/book-studio-logo.svg`
- Ícone da aplicação e favicon: `public/icon.svg`
- Manifesto PWA: `public/site.webmanifest`
- Imagem Open Graph: `src/app/opengraph-image.svg`
- Metadados globais, Open Graph, Twitter Cards, robots, ícones e tema: `src/app/layout.tsx`

## Stack
- Next.js App Router, React e TypeScript
- Tailwind CSS
- Firebase Authentication (integração planeada)
- Neon PostgreSQL como base de dados principal
- Google Drive API para armazenamento autorizado pelo utilizador
- NVIDIA API para modelos de texto e imagem
- MongoDB como adaptador opcional secundário para VPS/autohospedagem

## Desenvolvimento
Requer Node.js 22+ e pnpm.

1. Copie `.env.example` para `.env.local`.
2. Execute `pnpm install`.
3. Execute `pnpm dev`.

## Segurança e estado das integrações
A autenticação de produção deve usar Firebase Authentication e autorização no servidor. Qualquer acesso temporário para revisão visual deve ficar restrito ao ambiente de desenvolvimento/pré-visualização, nunca ser tratado como autenticação real nem ficar ativo em produção. As integrações só devem indicar sucesso após confirmação do serviço externo.

## Documentação
Consulte `docs/PRODUCT_SPEC.md`, `docs/ARCHITECTURE.md`, `docs/ROUTES.md` e `docs/DEVELOPMENT_STATUS.md`.

A licença open source ainda precisa de ser escolhida antes de uma publicação formal.
