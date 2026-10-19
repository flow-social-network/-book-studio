# Arquitetura do Book Studio

## Decisão arquitetural

O Book Studio será organizado como um monorepositório com aplicações separadas e contratos explícitos entre elas. A interface não implementa regras de negócio nem acessa o banco de dados diretamente: ela chama a API do backend. Componentes são reutilizáveis; páginas são finas e responsáveis por composição e roteamento.

## Estrutura-alvo

```text
/
├── web/                         # Site público completo (Next.js)
│   ├── src/app/                 # Rotas e layouts do site
│   ├── src/components/          # Componentes de UI, layout e funcionalidades
│   ├── src/styles/              # Estilos por componente/feature
│   ├── public/
│   └── package.json
├── auth/                        # Fluxos de identidade e acesso
│   ├── login/
│   ├── cadastro/
│   ├── recuperar-senha/
│   └── redefinir-senha/
├── backend/                     # API e toda a lógica de negócio
│   ├── cmd/api/
│   ├── internal/
│   │   ├── domain/
│   │   ├── application/
│   │   ├── ports/
│   │   ├── adapters/
│   │   ├── http/
│   │   └── infrastructure/
│   ├── migrations/
│   ├── tests/
│   └── go.mod
├── contracts/                   # Contratos OpenAPI e modelos públicos da API
├── docs/                        # Arquitetura, operação, segurança e decisões
├── scripts/
├── docker-compose.yml
└── README.md
```

> Durante a migração, o código Next.js atualmente em `src/app` não deve ser apagado nem movido de forma massiva sem atualizar imports, configuração de build, testes e deploy. A migração para `web/` será incremental e validada em cada etapa.

## Responsabilidades

### `web/`
- Site público, páginas institucionais e área de aplicação.
- Componentes reutilizáveis em pastas próprias, com estilos próximos ao componente.
- Páginas com composição e roteamento; sem regras de domínio.
- Consumo da API exclusivamente através de um cliente HTTP tipado.
- Sem acesso direto ao PostgreSQL/Neon, MongoDB ou credenciais privadas.
- Sem decisões de autorização: o backend valida identidade, propriedade e permissões.

### `auth/`
- Interface de login, cadastro, recuperação e redefinição de senha.
- Os fluxos de autenticação podem usar Firebase Auth no cliente para as operações suportadas; tokens são enviados à API.
- O backend valida os tokens e aplica autorização por recurso.
- Não duplicar telas de autenticação: os fluxos devem importar os mesmos componentes de formulário e layout quando compartilhados.

### `backend/`
- Fonte de verdade das regras de negócio.
- Validação de entrada, autorização, limites, persistência, auditoria e integrações externas.
- APIs versionadas, contratos tipados e erros explícitos.
- Nenhum endpoint declara sucesso se a operação persistente ou externa não foi confirmada.
- Segredos e credenciais ficam somente no ambiente do servidor.

## Tecnologia proposta

- **Web:** Next.js App Router, React e TypeScript, preservando a base existente.
- **Backend:** Go, como serviço independente, com camadas de domínio/aplicação/adaptadores e APIs HTTP versionadas. Go é adequado para um backend separado e serviços de longa duração; a escolha não elimina a necessidade de implementar e testar os adaptadores reais.
- **Banco principal:** Neon PostgreSQL.
- **Autenticação:** Firebase Authentication, com validação dos tokens no backend.
- **Arquivos:** integração autorizada com Google Drive; metadados e permissões permanecem sob controle da API.
- **IA:** adaptador de provedor NVIDIA configurado exclusivamente no backend.
- **Banco alternativo:** MongoDB como opção secundária para instalações autohospedadas, atrás de uma interface de persistência; não deve criar dois estados concorrentes nem ser ativado silenciosamente.

A compatibilidade atual do Firebase Admin SDK para Go e o suporte do Go a PostgreSQL são documentados oficialmente; as integrações concretas ainda precisam ser configuradas com credenciais reais e testadas. Referências: https://firebase.google.com/docs/admin/setup e https://go.dev/doc/database/.

## Regras de componentes

1. Cada componente tem uma responsabilidade e uma API de props tipada.
2. Reutilizar componentes existentes antes de criar novos.
3. CSS global apenas para tokens, reset, tipografia e regras realmente globais.
4. Estilos de componente/feature ficam próximos do componente, sem seletores globais acidentais.
5. Não misturar interface, acesso HTTP, validação de domínio e persistência.
6. Mudanças pequenas e localizadas; antes de substituir um arquivo, verificar dependências e consumidores.
7. Sem dados falsos apresentados como reais, fallbacks silenciosos ou botões sem ação funcional.

## Contrato de comunicação

- Base URL configurável por ambiente, por exemplo `BOOK_STUDIO_API_URL` no servidor e uma URL pública explicitamente configurada para chamadas do navegador.
- API versionada em `/api/v1`.
- Autenticação por token validado pelo backend.
- Respostas de erro consistentes, com código, mensagem segura e identificador de correlação.
- OpenAPI em `contracts/openapi.yaml` como contrato entre web e backend.
- Timeout, tratamento de indisponibilidade e estados de carregamento/erro explícitos.
- CORS restrito aos domínios configurados; nunca usar origem universal com credenciais.

## Ordem segura de implementação

1. Inventariar rotas, componentes, estilos, integrações e dependências existentes.
2. Criar a API Go com configuração validada, health/readiness e observabilidade.
3. Definir contratos OpenAPI e testes de contrato.
4. Implementar conexão real ao Neon, migrações e repositórios com isolamento por usuário.
5. Implementar validação Firebase e autorização no backend.
6. Construir casos de uso de projetos, livros, capítulos e conteúdo editorial.
7. Criar cliente HTTP tipado em `web/` e migrar uma funcionalidade de cada vez.
8. Separar as páginas de autenticação em `auth/` sem duplicar componentes.
9. Integrar Drive e NVIDIA por adaptadores de backend, com erros e limites explícitos.
10. Validar build, tipos, lint, testes unitários, integração, isolamento entre usuários e fluxo ponta a ponta antes de declarar pronto.

## Critérios de aceite

- Nenhuma página fala diretamente com o banco.
- Nenhuma chave privada de provedor é enviada ao navegador.
- Um usuário não consegue ler ou alterar projetos de outro usuário, mesmo conhecendo o ID.
- Os dados continuam disponíveis depois de atualizar a página e entrar novamente.
- Falhas de API/banco/provedor são visíveis e não viram sucesso fictício.
- Os fluxos de login, cadastro e recuperação usam autenticação real.
- Build e testes são executados no ambiente do projeto e seus resultados são registrados.
