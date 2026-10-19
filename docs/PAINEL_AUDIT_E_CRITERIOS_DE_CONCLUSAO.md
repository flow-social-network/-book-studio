# Auditoria de completude do painel

Data da auditoria: 2026-10-19.

## Regra de aceite

Uma rota listada no menu não está concluída só porque existe um `page.tsx`. Para considerar uma funcionalidade concluída, deve ter componente de feature reutilizável, contrato de API, validação no backend, persistência real quando aplicável, estados de carregamento/erro/vazio reais e testes. Não aceitar dados mockados, demos, localStorage como banco de projetos, nem mensagens de sucesso sem confirmação da API.

## Rotas atualmente identificadas

| Rota | Área | Situação observada |
|---|---|---|
| `/app` | Visão geral | Interface existente; precisa consultar contagens/atividade reais da API. |
| `/app/projetos` | Meus projetos | Usa `localStorage`; deve ser migrada para API autenticada e persistência real. |
| `/app/projetos/novo` | Criar livro | Salva rascunho no navegador; deve criar projeto real e redirecionar para o editor por ID. |
| `/app/editor` | Editor | Página de orientação; editor funcional não implementado. |
| `/app/editor/[projectId]` | Editor de projeto | Rota dedicada ainda precisa ser criada. |
| `/app/biblioteca` | Biblioteca visual | Estado vazio honesto; upload, metadados e armazenamento ainda não implementados. |
| `/app/ia` | Assistente editorial | Cartões informativos; chamada real ao provedor e contexto do manuscrito ainda não implementados. |
| `/app/publicacao` | Diagramação/exportação | Links externos; geradores/validação de exportação não implementados. |
| `/app/editor/estrutura` | Estrutura do livro | Página informativa; CRUD de partes/capítulos ainda não implementado. |
| `/app/editor/revisao` | Revisão editorial | Página informativa; análises e sugestões persistidas ainda não implementadas. |
| `/app/editor/previsualizacao` | Pré-visualização | Página informativa; motor de pré-visualização ainda não implementado. |
| `/app/perfil` | Perfil | Sem identidade autenticada conectada. |
| `/app/configuracoes` | Configurações | Navegação existente; preferências e configurações precisam de persistência/estado real. |
| `/app/configuracoes/preferencias` | Preferências | Deve ser auditada quanto à persistência no usuário autenticado. |
| `/app/integracoes` | Integrações | Lista informativa; estados devem vir de verificações reais do backend. |
| `/app/integracoes/google-drive` | Google Drive | OAuth e sincronização não implementados. |
| `/app/ajuda` | Ajuda | Conteúdo estático; validar links e documentação. |

## Padrão obrigatório para cada rota

1. `page.tsx` apenas compõe o componente da feature e resolve parâmetros/roteamento.
2. Interface e comportamento ficam em `src/components/features/<feature>/` (ou na estrutura equivalente após a migração para `web/`).
3. Elementos partilhados ficam em `src/components/ui/` e `src/components/layout/`.
4. Chamadas HTTP passam por um cliente tipado único; nenhum componente acessa diretamente o banco.
5. Erros de API são apresentados explicitamente. Nunca converter falha de rede em lista vazia ou sucesso.
6. Estado vazio só é apresentado depois de uma resposta bem-sucedida que confirme zero resultados.
7. Toda mutação exige resposta real confirmada pelo backend; não usar `localStorage` para simular persistência.
8. As ações devem respeitar a identidade e a autorização verificadas no servidor.

## Ordem de execução

1. Estabelecer cliente HTTP comum e contratos de API.
2. Implementar backend Go, autenticação e persistência PostgreSQL/Neon.
3. Concluir o fluxo de projetos: criar, listar, abrir por ID, atualizar e excluir com autorização.
4. Concluir o editor real com conteúdo estruturado e salvamento persistente.
5. Concluir estrutura, revisão, biblioteca visual, IA, pré-visualização e exportação ligadas ao mesmo projeto.
6. Concluir perfil, preferências, notificações e estado de integrações.
7. Migrar a estrutura para `web/` e separar os fluxos de identidade em `auth/` sem duplicar componentes.
8. Executar typecheck, lint, build, testes de integração e fluxo ponta a ponta.

## Critério final

A auditoria só pode ser encerrada quando cada rota do menu e cada ação visível estiver ligada a um fluxo real, tiver testes e não contiver dados de demonstração, persistência local substitutiva ou estados de sucesso fictícios.
