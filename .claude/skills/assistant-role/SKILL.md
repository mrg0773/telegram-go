---
name: assistant-role
description: Правила работы ассистента для @mrg0773. Фокус на деплой и CI/CD, минимум лишних действий.
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# Роль ассистента

## Основные правила

1. **Фокус на деплой и CI/CD** - занимайся только тем, что просят
2. **Не отвлекайся на issues** - если не просят, не смотри
3. **Не удаляй локальные директории** - только с GitLab
4. **Читай инструкции внимательно** - различай `users` и `tgferma-users`
5. **Минимум слов** - делай, не объясняй

## Что НЕ делать

- Не предлагай "постепенную миграцию"
- Не проверяй issues без запроса
- Не удаляй локальные файлы/папки
- Не добавляй лишние проверки
- Не объясняй что будешь делать - просто делай

## Типичные задачи

### Деплой пакета (Core Libraries)

**ВАЖНО**: Для @mrg0773 библиотек НЕ нужны теги и npm version!
CI автоматически увеличивает версию при пуше в main.

```bash
cd ~/Cursor/{lib-name}
git add .
git commit -m "fix: описание"
git push origin main
# CI автоматически: build → auto-increment version → publish
```

### Проверка пайплайна
```bash
curl -s -L -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
  "https://git.tgfermagames.ru/api/v4/projects/{path}/pipelines?per_page=3"
```

### Удаление проекта с GitLab
```bash
curl -s -X DELETE -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
  "https://git.tgfermagames.ru/api/v4/projects/{ID}"
```

### Восстановление проекта
```bash
curl -s -X POST -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
  "https://git.tgfermagames.ru/api/v4/projects/{ID}/restore"
```

## Project IDs (часто используемые)

| ID | Project |
|----|---------|
| 50 | serverless-function-lib |
| 53 | ydb-tgferma |
| 69 | users |
| 77 | ydb |

## Токен

```bash
export GITLAB_TOKEN="glpat-HV7Naqff3OFxMJc2ToxrxW86MQp1OjIH.01.0w0wh65th"
```
