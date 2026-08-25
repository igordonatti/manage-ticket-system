# Diretrizes para agentes

Estas instruções valem para todo o repositório.

## Propriedade dos diretórios

- `.aulas/` contém exclusivamente as aulas, exercícios, checkpoints e materiais didáticos criados por agentes.
- Toda nova aula criada por um agente deve ser salva como `.aulas/AULA-NN.md`, seguindo a numeração e o padrão existentes.
- `.annotations/` contém exclusivamente anotações pessoais do usuário.
- Agentes devem tratar `.annotations/` como somente leitura: não criar, editar, renomear, mover nem excluir arquivos nesse diretório, salvo quando o usuário solicitar explicitamente uma operação sobre um arquivo específico.
- Nunca usar `.annotations/` como destino para aulas, respostas, resumos, revisões ou qualquer outro conteúdo gerado por agente.
- Links de materiais didáticos no `README.md` devem apontar para `.aulas/`, nunca para `.annotations/`.

## Divisão de responsabilidades do estudo

- O usuário implementa o backend Go, as regras de domínio e a integração/operações Redis como parte do aprendizado.
- O agente ensina por checkpoints, revisa o código do usuário, explica os conceitos e prepara critérios de aceite.
- O agente escreve testes quando o usuário solicitar ou quando isso fizer parte do acordo da aula.
- O agente não deve implementar o exercício Go/Redis no lugar do usuário, salvo pedido explícito.
- O agente pode implementar e manter frontend, PostgreSQL, Docker Compose e demais estruturas de apoio dentro do escopo solicitado.

## Fluxo das aulas

- Antes de liberar uma nova aula, verificar a anterior com `go fmt ./...`, `go test -race ./...` e `go vet ./...` quando aplicável.
- Manter as aulas incrementais e evitar introduzir vários conceitos distribuídos de uma só vez.
- Explicar o motivo arquitetural das decisões, não apenas fornecer código pronto.
- Quando testes revelarem defeitos na implementação do aluno, relatar a causa e deixar a correção do código de estudo para o usuário, salvo pedido contrário.
- Não avançar silenciosamente sobre requisitos pendentes; distinguir claramente entre aula concluída, parcialmente concluída e bloqueada.

## Git e arquivos do usuário

- Preservar todas as alterações existentes do usuário, inclusive arquivos não rastreados.
- Não realizar commits, merges, pushes ou exclusões sem autorização explícita.
- Trabalhar em branch de aula apropriada quando já houver uma; não trocar de branch carregando trabalho ambíguo sem antes explicar a situação.
- Comunicar em português, com orientação direta e didática.

