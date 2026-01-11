---
name: serverless-yandex
description: Yandex Cloud Serverless. Cloud Functions, Containers, API Gateway, триггеры. Используй для деплоя функций, настройки триггеров, работы с Yandex Cloud.
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# Yandex Cloud Serverless

## Cloud Functions

### Структура функции

```typescript
// index.ts
import { Handler } from '@yandex-cloud/function-types';

export const handler: Handler.Http = async (event, context) => {
  // event - входящий запрос
  // context - контекст выполнения

  return {
    statusCode: 200,
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ message: 'Hello!' })
  };
};
```

### Event Types

```typescript
// HTTP trigger
interface HttpEvent {
  httpMethod: string;
  headers: Record<string, string>;
  queryStringParameters: Record<string, string>;
  body: string;
  isBase64Encoded: boolean;
}

// Timer trigger
interface TimerEvent {
  trigger_id: string;
  timer_id: string;
}

// Message Queue trigger
interface MessageQueueEvent {
  messages: Array<{
    event_metadata: object;
    body: string;
  }>;
}

// Object Storage trigger
interface ObjectStorageEvent {
  bucket_id: string;
  object_id: string;
  event_type: string;
}
```

### Context

```typescript
interface Context {
  requestId: string;
  functionName: string;
  functionVersion: string;
  memoryLimitInMB: number;
  token?: IAMToken;

  // Время до таймаута
  getRemainingTimeInMillis(): number;
}
```

## Деплой

### Через CLI (yc)

```bash
# Создать функцию
yc serverless function create --name my-function

# Создать версию
yc serverless function version create \
  --function-name my-function \
  --runtime nodejs18 \
  --entrypoint index.handler \
  --memory 128m \
  --execution-timeout 10s \
  --source-path ./dist

# Посмотреть логи
yc serverless function logs my-function --since 1h
```

### Через Terraform

```hcl
resource "yandex_function" "my_function" {
  name               = "my-function"
  user_hash          = filemd5("./dist/index.js")
  runtime            = "nodejs18"
  entrypoint         = "index.handler"
  memory             = "128"
  execution_timeout  = "10"

  content {
    zip_filename = "function.zip"
  }

  environment = {
    YDB_DATABASE = var.ydb_database
  }

  service_account_id = var.service_account_id
}
```

## API Gateway

### OpenAPI спецификация

```yaml
openapi: 3.0.0
info:
  title: My API
  version: 1.0.0

paths:
  /users/{id}:
    get:
      x-yc-apigateway-integration:
        type: cloud_functions
        function_id: ${function_id}
        service_account_id: ${service_account_id}
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Success

  /static/{file+}:
    get:
      x-yc-apigateway-integration:
        type: object_storage
        bucket: my-bucket
        object: '{file}'
```

## Триггеры

### Timer (Cron)

```bash
yc serverless trigger create timer \
  --name hourly-task \
  --cron-expression "0 * * * *" \
  --invoke-function-name my-function \
  --invoke-function-service-account-id $SA_ID
```

### Message Queue

```bash
yc serverless trigger create message-queue \
  --name queue-trigger \
  --queue arn:yc:ymq:ru-central1:xxx:my-queue \
  --queue-service-account-id $SA_ID \
  --invoke-function-name my-function \
  --invoke-function-service-account-id $SA_ID \
  --batch-size 10 \
  --batch-cutoff 10s
```

### Object Storage

```bash
yc serverless trigger create object-storage \
  --name storage-trigger \
  --bucket-id my-bucket \
  --events create-object,update-object \
  --invoke-function-name my-function \
  --invoke-function-service-account-id $SA_ID
```

## Serverless Containers

### Dockerfile

```dockerfile
FROM node:22-alpine

WORKDIR /app
COPY package*.json ./
RUN npm ci --production
COPY dist ./dist

EXPOSE 8080
CMD ["node", "dist/index.js"]
```

### Деплой контейнера

```bash
# Собрать и запушить образ
docker build -t cr.yandex/$REGISTRY_ID/my-app:latest .
docker push cr.yandex/$REGISTRY_ID/my-app:latest

# Создать контейнер
yc serverless container create --name my-container

# Создать ревизию
yc serverless container revision deploy \
  --container-name my-container \
  --image cr.yandex/$REGISTRY_ID/my-app:latest \
  --memory 512M \
  --execution-timeout 30s \
  --service-account-id $SA_ID
```

## IAM и авторизация

### Service Account

```bash
# Создать SA
yc iam service-account create --name my-sa

# Назначить роли
yc resource-manager folder add-access-binding $FOLDER_ID \
  --role serverless.functions.invoker \
  --subject serviceAccount:$SA_ID

yc resource-manager folder add-access-binding $FOLDER_ID \
  --role ydb.editor \
  --subject serviceAccount:$SA_ID
```

### IAM Token в функции

```typescript
// В Cloud Functions токен доступен автоматически
const iamToken = context.token?.access_token;

// Или через metadata
const response = await fetch(
  'http://169.254.169.254/computeMetadata/v1/instance/service-accounts/default/token',
  { headers: { 'Metadata-Flavor': 'Google' } }
);
const { access_token } = await response.json();
```

## Секреты

### Lockbox

```bash
# Создать секрет
yc lockbox secret create \
  --name my-secrets \
  --payload '[{"key": "DB_PASSWORD", "text_value": "secret123"}]'

# Использовать в функции
yc serverless function version create \
  --function-name my-function \
  --secrets-lockbox-id $SECRET_ID \
  ...
```

### В коде

```typescript
// Секреты доступны как переменные окружения
const dbPassword = process.env.DB_PASSWORD;
```

## Логирование

```typescript
// Логи идут в Yandex Cloud Logging
console.log('Info message');
console.error('Error message');

// Структурированные логи
console.log(JSON.stringify({
  level: 'info',
  message: 'User created',
  userId: '123'
}));
```

## Мониторинг

```bash
# Метрики функции
yc monitoring metric list \
  --service serverless-functions \
  --folder-id $FOLDER_ID

# Важные метрики:
# - invocations_count
# - errors_count
# - duration_ms
# - throttled_count
```

## Best Practices

1. **Cold start** - минимизируй зависимости
2. **Memory** - начинай с 128MB, увеличивай при необходимости
3. **Timeout** - ставь реалистичный (не max)
4. **Secrets** - используй Lockbox, не env vars
5. **Logging** - структурированные JSON логи
6. **Retry** - idempotent операции
7. **Connection pooling** - переиспользуй соединения между вызовами
