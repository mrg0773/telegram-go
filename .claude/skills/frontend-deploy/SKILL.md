---
name: frontend-deploy
description: Деплой статических фронтенд-сайтов в Yandex Cloud. S3 bucket, AGW с CORS, CI/CD pipeline. Используй для настройки деплоя React/Vue/Vite приложений.
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# Frontend Static Site Deployment

## Быстрый старт

### 1. Создать S3 bucket

```bash
# Создать bucket с website hosting
yc storage bucket create \
  --name my-frontend \
  --default-storage-class standard \
  --acl public-read

yc storage bucket update \
  --name my-frontend \
  --website-settings '{"index": "index.html", "error": "index.html"}'
```

### 2. GitLab CI Variables

| Variable | Value | Protected |
|----------|-------|-----------|
| `YC_S3_ACCESS_KEY` | Access Key ID | Yes |
| `YC_S3_SECRET_KEY` | Secret Access Key | Yes |
| `S3_BUCKET` | my-frontend | No |
| `S3_ENDPOINT` | https://storage.yandexcloud.net | No |
| `DEPLOY_URL` | https://my-frontend.website.yandexcloud.net | No |

### 3. .gitlab-ci.yml (минимальный)

```yaml
stages:
  - deploy

deploy:
  stage: deploy
  image: node:22-alpine
  before_script:
    - apk add --no-cache aws-cli
    - export AWS_ACCESS_KEY_ID=$YC_S3_ACCESS_KEY
    - export AWS_SECRET_ACCESS_KEY=$YC_S3_SECRET_KEY
    - export AWS_DEFAULT_REGION=ru-central1
  script:
    - npm ci
    - npm run build
    # ВАЖНО: Явно указываем content-type для JS/CSS файлов
    # Без этого S3 отдает text/plain и браузер блокирует ES модули
    - |
      for file in $(find dist/assets -name "*.js" 2>/dev/null); do
        aws s3 cp "$file" "s3://${S3_BUCKET}/assets/$(basename $file)" \
          --endpoint-url=$S3_ENDPOINT --acl public-read \
          --content-type "application/javascript"
      done
    - |
      for file in $(find dist/assets -name "*.css" 2>/dev/null); do
        aws s3 cp "$file" "s3://${S3_BUCKET}/assets/$(basename $file)" \
          --endpoint-url=$S3_ENDPOINT --acl public-read \
          --content-type "text/css"
      done
    # Остальные файлы
    - aws s3 sync dist/ s3://$S3_BUCKET/ --endpoint-url=$S3_ENDPOINT --delete --acl public-read --exclude "assets/*.js" --exclude "assets/*.css"
  only:
    - main
```

## CORS Configuration

### Вариант 1: CORS на бакете

```bash
cat > cors.json << 'EOF'
{
  "cors_rules": [{
    "allowed_methods": ["GET", "HEAD"],
    "allowed_origins": ["*"],
    "allowed_headers": ["*"],
    "max_age_seconds": 3600
  }]
}
EOF

yc storage bucket update --name my-frontend --cors cors.json
```

### Вариант 2: CORS через AGW (рекомендуется)

```yaml
# openapi.yaml для AGW
openapi: 3.0.0
info:
  title: Frontend
  version: 1.0.0

x-yc-apigateway:
  cors:
    origin: '*'
    methods: 'GET, HEAD, OPTIONS'
    headers: '*'
    maxAge: 86400

paths:
  /{path+}:
    get:
      x-yc-apigateway-integration:
        type: object_storage
        bucket: my-frontend
        object: '{path}'
        error_object: 'index.html'
  /:
    get:
      x-yc-apigateway-integration:
        type: object_storage
        bucket: my-frontend
        object: 'index.html'
```

```bash
yc serverless api-gateway create --name my-frontend-gw --spec=openapi.yaml
```

## Cache Strategy

```yaml
# CI/CD script
# Assets (хешированные) - долгий кэш
aws s3 sync dist/assets/ s3://$BUCKET/assets/ \
  --cache-control "public, max-age=31536000, immutable" \
  --exclude "*.map"

# HTML - без кэша
aws s3 sync dist/ s3://$BUCKET/ \
  --cache-control "public, max-age=0, must-revalidate" \
  --exclude "assets/*"
```

## URL Types

| Type | Format | Use Case |
|------|--------|----------|
| S3 Website | `https://{bucket}.website.yandexcloud.net` | Dev/Test |
| AGW | `https://{id}.apigw.yandexcloud.net` | Prod с CORS |
| Custom | `https://app.example.com` | Production |

## DEPLOY.md Template

Каждый frontend репозиторий должен содержать `DEPLOY.md`:

```markdown
# Deployment

## URLs
- Production: https://...
- AGW: https://...

## Infrastructure
- S3 Bucket: `bucket-name`
- AGW ID: `d5d...`

## Manual Deploy
export AWS_ACCESS_KEY_ID="..."
export AWS_SECRET_ACCESS_KEY="..."
npm run deploy
```

## Troubleshooting

| Проблема | Причина | Решение |
|----------|---------|---------|
| 403 Forbidden | Bucket не public | `--acl public-read` |
| CORS error | Нет CORS config | Настроить CORS на bucket или AGW |
| 404 на routes | Нет error document | `error: index.html` в website-settings |
| Старый кэш | HTML закэширован | `max-age=0` для HTML |
| **MIME type error** | S3 отдает JS как text/plain | `--content-type "application/javascript"` при загрузке |
| CSS не загружается | S3 отдает CSS как text/plain | `--content-type "text/css"` при загрузке |

### TypeError: 'text/plain' is not a valid JavaScript MIME type

Браузеры блокируют ES модули с неправильным MIME типом. AWS CLI в Docker не имеет `/etc/mime.types` и не может автоматически определить тип файла.

**Решение**: Загружать JS/CSS файлы с явным `--content-type` (см. CI/CD пример выше).
