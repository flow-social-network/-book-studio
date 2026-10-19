# Book Studio Backend

Serviço de API independente. O backend é a fonte de verdade das regras de negócio, autorização, persistência e integrações. A implementação deve seguir `docs/ARCHITECTURE_MONOREPO.md`.

## Estado

Este diretório é o ponto de entrada da migração para um backend Go. A API ainda não deve ser considerada pronta para produção até que configuração, autenticação Firebase, persistência Neon, migrações, contratos, testes e integração com o frontend estejam implementados e validados.

## Requisitos

- Go (versão definida no `go.mod` quando o serviço for inicializado).
- Neon PostgreSQL configurado por segredo de ambiente.
- Firebase Auth configurado para validação de tokens.
- Segredos nunca devem ser commitados.

## Princípios

- API versionada em `/api/v1`.
- Sem mocks ou fallbacks silenciosos em produção.
- Isolamento de dados por usuário validado no servidor.
- Health e readiness separados: readiness só fica saudável quando dependências obrigatórias estiverem disponíveis.
- Testes antes de declarar qualquer fluxo concluído.
