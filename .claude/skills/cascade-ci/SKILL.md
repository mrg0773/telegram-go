---
name: cascade-ci
description: Каскадное обновление и CI/CD для @mrg0773 библиотек. Используй когда нужно обновить все библиотеки, запустить CI, опубликовать пакет, или разобраться с pipeline.
allowed-tools: Read, Grep, Glob, Bash
---

# Каскадное CI/CD

## Что это

Система автоматического обновления всех 22 библиотек в правильном порядке зависимостей.

## Запуск

### Локально

```bash
cd ~/Cursor/serverless-function-lib
./scripts/cascade-update.sh
```

### Через GitLab

1. Открыть CI/CD → Pipelines
2. Run pipeline
3. Добавить переменную: `CASCADE_UPDATE=true`
4. Run

## Порядок обновления

Pipeline обновляет библиотеки в порядке зависимостей:

```
1. Base Layer (параллельно):
   logger-error-lib, trace-lib, schemas, types, redis

2. YDB Layer (последовательно):
   ydb → ydb-toolkit → ydb-apps, ydb-users, ydb-bots

3. Core Layer (последовательно):
   ydb-tgferma → tgferma → tgferma-users, tgferma-admin

4. External Layer (параллельно):
   c2d, telegram-lib, openai-lib
```

## Что делает cascade-update

Для каждой библиотеки:

1. **npm update @mrg0773/*** - обновляет зависимости
2. **npm run build** - собирает
3. **npm run lint** - проверяет код
4. **git commit** - коммитит изменения
5. **npm version patch** - поднимает версию
6. **git push --tags** - пушит с тегом
7. Ждёт публикации в GitHub Packages

При ошибке - создаёт issue в GitLab.

## Публикация одной библиотеки

**ВАЖНО**: Теги и npm version НЕ нужны! CI делает auto-increment автоматически.

```bash
cd ~/Cursor/{lib-name}

# Просто пуш в main
git add .
git commit -m "fix: описание изменений"
git push origin main
```

Pipeline автоматически:
1. Соберёт проект (npm run build)
2. Проверит опубликованную версию
3. Увеличит patch если версия уже опубликована
4. Исправит пути в package.json (уберёт dist/)
5. Опубликует в GitHub Packages

### Принудительное обновление версии (редко)

Если нужно поднять minor/major версию:
```bash
# Обновить версию в package.json вручную
# например: "version": "2.0.0"
git add package.json
git commit -m "feat: major update"
git push origin main
```

## CI/CD переменные в GitLab

| Переменная | Назначение |
|------------|------------|
| NPM_TOKEN | Чтение пакетов из GitHub Packages |
| NPM_WRITE_TOKEN | Публикация в GitHub Packages |
| GITLAB_TOKEN | Создание issues при ошибках |
| CASCADE_UPDATE | Триггер каскадного обновления |

## Структура Pipeline

```yaml
stages:
  - build      # Сборка + линтер
  - publish    # Публикация (только по тегу)
  - cascade    # Каскадное обновление (опционально)
```

## При ошибке build

1. Автоматически создаётся issue в GitLab
2. Получаешь уведомление
3. Исправь ошибки и запуши
4. Pipeline перезапустится

## Проверка статуса библиотек

```bash
cd ~/Cursor/serverless-function-lib
./scripts/visualize-deps.sh
```

## Файлы конфигурации

- `.gitlab-ci.yml` - CI/CD конфигурация
- `scripts/cascade-update.sh` - скрипт каскадного обновления
- `scripts/visualize-deps.sh` - статус библиотек

## Troubleshooting

### Pipeline завис
```bash
# Проверить статус
./scripts/visualize-deps.sh

# Перезапустить вручную
cd ~/Cursor/{lib-name}
npm run build && git push
```

### Ошибка публикации
```bash
# Проверить .npmrc
cat .npmrc
# Должно быть:
# @mrg0773:registry=https://npm.pkg.github.com
# //npm.pkg.github.com/:_authToken=${NPM_TOKEN}

# Проверить версию (не должна быть уже опубликована)
npm view @mrg0773/{lib-name} versions
```

### Конфликт версий
```bash
# Посмотреть что требует какую версию
npm ls @mrg0773/{dep-name}

# Обновить зависимость
npm update @mrg0773/{dep-name}
```
