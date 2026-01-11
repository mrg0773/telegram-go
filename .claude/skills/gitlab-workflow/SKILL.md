---
name: gitlab-workflow
description: Работа с GitLab self-hosted для @mrg0773. Используй для создания issues, merge requests, работы с CI/CD, публикации пакетов. GitLab на git.tgfermagames.ru.
allowed-tools: Read, Grep, Glob, Bash
---

# GitLab Workflow

## GitLab Self-Hosted

- **URL**: https://git.tgfermagames.ru
- **Группа**: @mrg0773
- **API**: https://git.tgfermagames.ru/api/v4

## Авторизация

Токен уже добавлен в `~/.zshrc`:

```bash
export GITLAB_TOKEN="glpat-HV7Naqff3OFxMJc2ToxrxW86MQp1OjIH.01.0w0wh65th"
```

Для новой сессии: `source ~/.zshrc`

## Issues

### Создать issue

```bash
curl -X POST \
  -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Название issue",
    "description": "Описание проблемы",
    "labels": "bug,priority::high"
  }' \
  "https://git.tgfermagames.ru/api/v4/projects/{PROJECT_ID}/issues"
```

### Найти project_id

```bash
# По пути репозитория
curl -s -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
  "https://git.tgfermagames.ru/api/v4/projects?search={repo-name}" | jq '.[0].id'
```

### Список issues

```bash
curl -s -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
  "https://git.tgfermagames.ru/api/v4/projects/{PROJECT_ID}/issues?state=opened" | \
  jq -r '.[] | "#\(.iid): \(.title)"'
```

### Закрыть issue

```bash
curl -X PUT \
  -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
  -d "state_event=close" \
  "https://git.tgfermagames.ru/api/v4/projects/{PROJECT_ID}/issues/{ISSUE_IID}"
```

## Labels (метки)

Стандартные метки для @mrg0773:

| Label | Цвет | Назначение |
|-------|------|------------|
| bug | red | Баги |
| feature | green | Новый функционал |
| refactor | blue | Рефакторинг |
| ci-failure | orange | Ошибки CI (автоматически) |
| neverthrow | purple | Миграция на neverthrow |
| question | yellow | Вопросы |
| priority::high | red | Высокий приоритет |
| priority::low | gray | Низкий приоритет |

## NPM Registry (GitHub Packages)

### Настройка .npmrc

```bash
# Для чтения
echo "@mrg0773:registry=https://npm.pkg.github.com" > .npmrc
echo "//npm.pkg.github.com/:_authToken=${NPM_TOKEN}" >> .npmrc

# Для публикации (в dist/)
echo "@mrg0773:registry=https://npm.pkg.github.com" > dist/.npmrc
echo "//npm.pkg.github.com/:_authToken=${NPM_WRITE_TOKEN}" >> dist/.npmrc
```

### Публикация пакета

**ВАЖНО**: Теги и npm version НЕ нужны! CI делает auto-increment.

```bash
# Просто пуш в main
git add .
git commit -m "fix: описание"
git push origin main
# CI автоматически: build → auto-increment → publish
```

### Ручная публикация (редко)
```bash
npm run build
cd dist && npm publish
```

### Проверка версий

```bash
# Все версии пакета
npm view @mrg0773/{package-name} versions

# Последняя версия
npm view @mrg0773/{package-name} version
```

## CI/CD Pipeline

### Статус pipeline

```bash
# Последние pipelines
curl -s -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
  "https://git.tgfermagames.ru/api/v4/projects/{PROJECT_ID}/pipelines?per_page=5" | \
  jq -r '.[] | "\(.id): \(.status) (\(.ref))"'
```

### Перезапуск pipeline

```bash
curl -X POST \
  -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
  "https://git.tgfermagames.ru/api/v4/projects/{PROJECT_ID}/pipelines/{PIPELINE_ID}/retry"
```

### Запуск с переменными

```bash
curl -X POST \
  -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
  -F "ref=main" \
  -F "variables[CASCADE_UPDATE]=true" \
  "https://git.tgfermagames.ru/api/v4/projects/{PROJECT_ID}/pipeline"
```

## Git Remote

### Проверить remote

```bash
git remote -v
# Должно быть git.tgfermagames.ru
```

### Добавить remote (если нет)

```bash
git remote add origin https://git.tgfermagames.ru/mrg0773/{repo-name}.git
```

## Merge Requests

### Создать MR

```bash
curl -X POST \
  -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "source_branch": "feature/my-feature",
    "target_branch": "main",
    "title": "feat: описание",
    "remove_source_branch": true
  }' \
  "https://git.tgfermagames.ru/api/v4/projects/{PROJECT_ID}/merge_requests"
```

## Project IDs

### Core Libraries (core_lib/)

| ID | Project | Описание |
|----|---------|----------|
| 50 | serverless-function-lib | Главный репозиторий |
| 47 | logger-error | Логирование |
| 85 | trace-lib | Трейсинг |
| 63 | schemas | JSON схемы |
| 87 | telegram-lib | Telegram API |
| 49 | bl-fsm | FSM библиотека |
| 48 | jwt-auth | JWT авторизация |
| 46 | open-api-agw | OpenAPI AGW |

### YDB Libraries

| ID | Project | Описание |
|----|---------|----------|
| 77 | ydb | Базовый YDB клиент |
| 75 | ydb-toolkit | YDB утилиты |
| 84 | ydb-apps | YDB приложения |
| 82 | ydb-users | YDB пользователи |
| 83 | ydb-bots | YDB боты |
| 53 | ydb-tgferma | YDB TGFerma |
| 54 | ydb-cdc | YDB CDC |

### TGFerma (tgferma/core/libs/)

| ID | Project | Описание |
|----|---------|----------|
| 62 | tgferma | Основная логика |
| 69 | users | Пользователи |
| 73 | tgferma-admin | Админка |
| 89 | c2d | Click-to-Dial |
| 91 | redis | Redis |
| 90 | types | Типы |
| 92 | open-ai | OpenAI |

### Services (tgferma/services/)

| ID | Project | Описание |
|----|---------|----------|
| 70 | users-function | Users Function |
| 66 | ferma-api | Ferma API |
| 67 | panel-v4-backend | Panel Backend |
| 52 | deploy-actions-v4 | Deploy Actions |

### Найти ID проекта

```bash
curl -s -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
  "https://git.tgfermagames.ru/api/v4/projects?search=NAME" | jq '.[0].id'
```

## Типичные сценарии

### Создать issue о баге

```bash
PROJECT_ID=50  # serverless-function-lib
curl -X POST \
  -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Bug: описание бага",
    "description": "## Шаги воспроизведения\n1. ...\n2. ...\n\n## Ожидаемое поведение\n\n## Фактическое поведение",
    "labels": "bug"
  }' \
  "https://git.tgfermagames.ru/api/v4/projects/$PROJECT_ID/issues"
```

### Найти все открытые issues в группе

```bash
curl -s -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
  "https://git.tgfermagames.ru/api/v4/groups/mrg0773/issues?state=opened" | \
  jq -r '.[] | "\(.project_id): #\(.iid) \(.title)"'
```
