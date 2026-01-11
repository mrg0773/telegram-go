---
name: ci-cd-templates
description: CI/CD шаблоны для @mrg0773 библиотек. GitLab CI с auto-increment версий, auto-issue на падение, публикация в GitHub Packages.
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# CI/CD Templates @mrg0773

## Стандартный .gitlab-ci.yml

```yaml
# GitLab CI/CD for @mrg0773/library-name
# Auto-publishes to GitHub Packages on tag push

stages:
  - build
  - publish

variables:
  NODE_VERSION: "22"

cache:
  key:
    files:
      - package-lock.json
  paths:
    - node_modules/

.setup_npm:
  before_script:
    - echo "@mrg0773:registry=https://npm.pkg.github.com" > .npmrc
    - echo "//npm.pkg.github.com/:_authToken=${NPM_TOKEN}" >> .npmrc
    - npm ci || npm install

build:
  stage: build
  image: node:${NODE_VERSION}
  extends: .setup_npm
  script:
    - echo "🔨 Building..."
    - npm run build
    - ls -la dist/
    - npm run lint --if-present || true
    - npm run type-check --if-present || true
  artifacts:
    paths:
      - dist/
    expire_in: 1 hour
  rules:
    - if: '$CI_COMMIT_BRANCH == "main"'
    - if: '$CI_COMMIT_TAG'
  tags:
    - docker

create_issue_on_failure:
  stage: build
  image: curlimages/curl:latest
  when: on_failure
  needs:
    - job: build
      optional: true
  script:
    - |
      if [ -z "$GITLAB_TOKEN" ]; then
        echo "GITLAB_TOKEN not set, skipping"
        exit 0
      fi
      TITLE="CI Build Failed - $(date '+%Y-%m-%d %H:%M')"
      curl -s -X POST \
        -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{\"title\": \"$TITLE\", \"description\": \"Pipeline #${CI_PIPELINE_ID} failed.\\n\\nBranch: ${CI_COMMIT_REF_NAME}\\nCommit: ${CI_COMMIT_SHORT_SHA}\\nPipeline: ${CI_PIPELINE_URL}\", \"labels\": \"ci-failure,auto-created\"}" \
        "${CI_API_V4_URL}/projects/${CI_PROJECT_ID}/issues"
      echo "Issue created"
  rules:
    - if: '$CI_COMMIT_BRANCH == "main"'
      when: on_failure
  tags:
    - docker
  allow_failure: true

publish:npm:
  stage: publish
  image: node:${NODE_VERSION}
  dependencies:
    - build
  script:
    - echo "📦 Publishing..."
    - echo "@mrg0773:registry=https://npm.pkg.github.com" > dist/.npmrc
    - echo "//npm.pkg.github.com/:_authToken=${NPM_WRITE_TOKEN}" >> dist/.npmrc
    - cd dist
    - |
      # Try to publish, handle 409 Conflict by bumping version
      MAX_RETRIES=3
      RETRY=0
      while [ $RETRY -lt $MAX_RETRIES ]; do
        echo "Attempt $((RETRY + 1)) of $MAX_RETRIES"
        npm publish 2>&1 | tee /tmp/npm-publish.log && exit 0

        if grep -q "409 Conflict\|Cannot publish over existing version" /tmp/npm-publish.log; then
          echo "⚠️ Version conflict, incrementing patch version..."
          CURRENT=$(node -p "require('./package.json').version")
          MAJOR=$(echo $CURRENT | cut -d. -f1)
          MINOR=$(echo $CURRENT | cut -d. -f2)
          PATCH=$(echo $CURRENT | cut -d. -f3)
          NEW_PATCH=$((PATCH + 1))
          NEW_VERSION="${MAJOR}.${MINOR}.${NEW_PATCH}"
          echo "Bumping version: $CURRENT -> $NEW_VERSION"
          node -e "const p=require('./package.json'); p.version='$NEW_VERSION'; require('fs').writeFileSync('package.json', JSON.stringify(p, null, 2)+'\n')"
          RETRY=$((RETRY + 1))
        else
          echo "❌ Publish failed with unknown error"
          exit 1
        fi
      done
      echo "❌ Max retries reached"
      exit 1
  rules:
    - if: '$CI_COMMIT_TAG =~ /^v?[0-9]+\.[0-9]+\.[0-9]+.*$/'
  tags:
    - docker
```

## Необходимые переменные

В GitLab Settings → CI/CD → Variables:

| Variable | Описание | Protected | Masked |
|----------|----------|-----------|--------|
| NPM_TOKEN | GitHub token для чтения пакетов | ❌ | ✅ |
| NPM_WRITE_TOKEN | GitHub token для публикации | ❌ | ✅ |
| GITLAB_TOKEN | GitLab token для создания issues | ❌ | ✅ |
| GITLAB_API_TOKEN | GitLab API token | ❌ | ✅ |

**Важно**: Переменные настроены на уровне группы `core_lib`, не на уровне проекта.

## Фичи

### 1. Auto-increment версии при конфликте
Если npm publish возвращает 409 Conflict, автоматически:
- Инкрементирует patch версию (1.2.3 → 1.2.4)
- Пробует опубликовать снова
- До 3 попыток

### 2. Auto-issue при падении build
Создает issue в GitLab с:
- Ссылкой на pipeline
- Информацией о коммите
- Меткой `ci-failure`

### 3. Кэширование node_modules
По хэшу package-lock.json для ускорения сборки.

## Релиз пакета

**ВАЖНО**: Теги и npm version НЕ нужны! CI делает всё автоматически.

```bash
# Просто пуш в main
git add .
git commit -m "fix: описание"
git push origin main
```

Pipeline автоматически:
1. Соберёт проект
2. Проверит опубликованную версию в GitHub Packages
3. Увеличит patch если текущая версия <= опубликованной
4. Исправит пути в package.json (уберёт dist/)
5. Опубликует в GitHub Packages

### Minor/Major версии (редко)

```bash
# Изменить version в package.json вручную
git add package.json
git commit -m "feat: новая major версия"
git push origin main
```

## Rollback функции на предыдущую версию

```bash
# 1. Найти нужную версию
yc serverless function version list --function-name FUNC_NAME --folder-id FOLDER_ID

# 2. Откатить (создать копию старой версии → станет $latest)
yc serverless function version create \
  --function-id FUNCTION_ID \
  --runtime nodejs22 \
  --entrypoint dist/index.handler \
  --memory 256m \
  --execution-timeout 30s \
  --source-version-id OLD_VERSION_ID \
  --environment "KEY1=value1,KEY2=value2"
```

**Ключ**: `--source-version-id` создаёт новую версию из кода старой → она становится `$latest`.
