# Claude Code Instructions

## Skills

Доступные skills в `.claude/skills/`:
- **gitlab-workflow** - Работа с GitLab (issues, MR, CI/CD)
- **dev-workflow** - Правила разработки, workflow
- **qa-checklist** - Чек-лист тестирования

## GitLab Self-Hosted

- **URL**: https://git.tgfermagames.ru
- **Token**: `$GITLAB_TOKEN` (в ~/.zshrc)

### Создать issue

```bash
curl -X POST \
  -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title": "Bug: описание", "labels": "bug"}' \
  "https://git.tgfermagames.ru/api/v4/projects/{PROJECT_ID}/issues"
```

### Найти project_id

```bash
curl -s -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
  "https://git.tgfermagames.ru/api/v4/projects?search=REPO_NAME" | jq '.[0].id'
```

## Правила

1. **Баги в @mrg0773 библиотеках** → создать issue, не workaround
2. **Деплой** → только через git push, CI делает деплой
3. **После решения issue** → сразу коммит и пуш

