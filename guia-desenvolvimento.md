# SDD — Software Design Document

> Documento de orientação de desenvolvimento e deploy do serviço de cadastro de
> dispositivos (POC). Define **onde** o código mora e **como** o time versiona e
> integra mudanças. Deve ser lido antes de qualquer commit.

## Repositório

- **Repositório oficial:** https://github.com/balmiza/poc_pm_tech.git
- **Modelo:** monorepo. Código, documentação (`docs/`) e artefatos de banco de
  dados (`db/`) convivem no mesmo repositório, evoluindo juntos no mesmo commit.

Toda mudança de regra de negócio deve carregar, no mesmo Pull Request, as
alterações de código e de documentação correspondentes, para que doc e código
nunca saiam de sincronia.

## Estratégia de branches (gitflow)

O fluxo de versionamento segue três níveis:

```
feature/nome-da-branch  ->  develop  ->  main
```

### Branches permanentes

| Branch | Propósito |
|--------|-----------|
| `main` | Branch estável. Reflete o que está (ou está apto a ser) implantado em produção. Só recebe código vindo de `develop`. |
| `develop` | Branch de integração. Concentra as features concluídas e validadas antes de promover para `main`. |

### Branches temporárias

| Padrão | Propósito |
|--------|-----------|
| `feature/nome-da-branch` | Desenvolvimento de uma funcionalidade ou correção específica. Criada a partir de `develop` e integrada de volta em `develop`. |

O nome da branch de feature deve ser descritivo e em kebab-case, por exemplo:
`feature/associar-dispositivo`, `feature/consultar-autorizacao`.

## Fluxo de trabalho

1. Atualizar a `develop` local com a remota.
2. Criar a branch de feature a partir de `develop`:
   `git checkout -b feature/nome-da-branch develop`.
3. Desenvolver e commitar na branch de feature.
4. Abrir Pull Request da feature para `develop`.
5. Após revisão e validação, fazer o merge em `develop`.
6. Quando `develop` estiver estável e pronta para release, abrir Pull Request de
   `develop` para `main`.
7. Após o merge em `main`, o código está apto ao deploy.

### Regras de integração

- Nenhum commit é feito diretamente em `main`.
- Nenhum commit é feito diretamente em `develop`; mudanças entram via Pull
  Request a partir de uma branch de feature.
- Uma feature só sobe para `main` depois de passar por `develop`.

## Inicialização do repositório (primeiro commit)

Como este é o primeiro commit do repositório, as branches base ainda não existem
e precisam ser criadas. Sequência recomendada:

1. Inicializar o repositório local e adicionar o remoto:
   ```
   git init
   git remote add origin https://github.com/balmiza/poc_pm_tech.git
   ```
2. Criar o primeiro commit na branch `main`:
   ```
   git add .
   git commit -m "chore: estrutura inicial do projeto"
   git branch -M main
   git push -u origin main
   ```
3. Criar a branch `develop` a partir de `main` e publicá-la:
   ```
   git checkout -b develop
   git push -u origin develop
   ```
4. A partir daqui, todo desenvolvimento começa em branches `feature/*` criadas a
   partir de `develop`.

> Após a inicialização, `main` e `develop` passam a ser as branches permanentes
> do fluxo descrito acima.

## Deploy

- O deploy é feito **a partir da branch `main`**.
- Apenas código que passou por `develop` e foi promovido para `main` é elegível
  para deploy.
- O detalhamento do ambiente de hospedagem (infraestrutura, banco, dependências
  externas) será descrito em `docs/architecture.md`, ainda a ser elaborado.