---
name: debugging
description: Отладка и мониторинг в @mrg0773. Логирование, трейсинг, Sentry, CloudWatch. Используй для отладки, логов, трассировки ошибок.
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# Debugging & Monitoring

## Логирование

### Базовое использование

```typescript
import { info, error, warn, debug } from '@mrg0773/logger-error';

// Уровни логирования
debug('Detailed debug info', { data });
info('Operation completed', { result });
warn('Something suspicious', { warning });
error('Operation failed', { error: err });
```

### Структурированные логи

```typescript
import { logger } from '@mrg0773/logger-error';

// С контекстом
logger.info('User created', {
  userId: '123',
  action: 'create',
  duration: 150
});

// Ошибка с контекстом
logger.error('Database query failed', {
  query: 'SELECT...',
  error: err.message,
  stack: err.stack
});
```

### Форматирование для CloudWatch

Логи автоматически форматируются для Yandex Cloud Functions:

```json
{
  "level": "info",
  "message": "User created",
  "timestamp": "2024-01-15T10:30:00.000Z",
  "context": {
    "userId": "123",
    "action": "create"
  }
}
```

## Distributed Tracing

### Trace Context

```typescript
import { TraceContext, generateTraceId } from '@mrg0773/trace-lib';

// Создать контекст
const traceId = generateTraceId();
const context = new TraceContext(traceId);

// Передать в сервисы
await userService.create(userData, context);
```

### Correlation ID

```typescript
// В handler
export async function handler(event: APIGatewayEvent) {
  // Получить или создать correlation ID
  const correlationId = event.headers['x-correlation-id'] || generateTraceId();

  // Добавить во все логи
  logger.info('Request started', { correlationId });

  // Передать в ответ
  return {
    statusCode: 200,
    headers: {
      'x-correlation-id': correlationId
    },
    body: JSON.stringify(result)
  };
}
```

### Middleware

```typescript
import { traceMiddleware } from '@mrg0773/trace-lib';

// Автоматическое добавление trace context
const handler = traceMiddleware(async (event, context) => {
  // context.traceId доступен везде
  logger.info('Processing', { traceId: context.traceId });
});
```

## Sentry Integration

### Настройка

```typescript
import * as Sentry from '@sentry/node';

Sentry.init({
  dsn: process.env.SENTRY_DSN,
  environment: process.env.ENVIRONMENT,
  tracesSampleRate: 0.1
});
```

### Отправка ошибок

```typescript
import { captureException, captureMessage } from '@sentry/node';

try {
  await riskyOperation();
} catch (error) {
  // Отправить в Sentry
  captureException(error, {
    tags: {
      operation: 'riskyOperation',
      userId: userId
    },
    extra: {
      input: inputData
    }
  });
}
```

### Performance Tracing

```typescript
import { createSentrySpan } from '@mrg0773/redis';

// Создать span для операции
const span = createSentrySpan('db.query', 'SELECT users');

try {
  const result = await db.query(...);
  span.setStatus('ok');
  return result;
} catch (error) {
  span.setStatus('error');
  throw error;
} finally {
  span.finish();
}
```

### Optional Sentry (в библиотеках)

```typescript
// Работает без инициализации Sentry
import { createSentrySpan, mockSpan } from '@mrg0773/redis';

// Если Sentry не инициализирован - возвращает mock
const span = createSentrySpan('operation', 'description');
// span.finish() безопасен даже без Sentry
```

## Error Analysis

### SQL Error Analyzer

```typescript
import { sqlErrorAnalyzer } from '@mrg0773/ydb';

const error = await ydb.query`SELECT * FROM nonexistent`;

if (error.isErr()) {
  const analysis = sqlErrorAnalyzer(error.error);
  console.log(analysis);
  // {
  //   isRetryable: false,
  //   category: 'schema',
  //   suggestion: 'Check table name'
  // }
}
```

### Error Context

```typescript
import { createAppError } from '@mrg0773/logger-error';

const error = createAppError('DATABASE_ERROR', 'Query failed', {
  context: {
    query: 'SELECT...',
    params: { id: 123 },
    duration: 5000
  },
  cause: originalError
});

// В логах будет полный контекст
logger.error('Operation failed', {
  code: error.code,
  message: error.message,
  context: error.context,
  cause: error.cause?.message
});
```

## Debugging Tips

### Локальная отладка

```bash
# Запуск с debug логами
DEBUG=* npm run dev

# Только определённые модули
DEBUG=ydb:* npm run dev
```

### Inspect mode

```bash
# Node.js inspector
node --inspect dist/index.js

# Открыть chrome://inspect в Chrome
```

### Console методы

```typescript
// Таблица для массивов
console.table(users);

// Время выполнения
console.time('operation');
await operation();
console.timeEnd('operation'); // operation: 150ms

// Группировка логов
console.group('User Operations');
console.log('Creating user...');
console.log('Sending email...');
console.groupEnd();
```

## Мониторинг в продакшене

### Health Check

```typescript
export async function healthCheck() {
  const checks = {
    ydb: await checkYDB(),
    redis: await checkRedis(),
    external: await checkExternalServices()
  };

  const healthy = Object.values(checks).every(c => c.ok);

  return {
    statusCode: healthy ? 200 : 503,
    body: JSON.stringify({
      status: healthy ? 'healthy' : 'unhealthy',
      checks
    })
  };
}
```

### Metrics

```typescript
// Подсчёт операций
const metrics = {
  requests: 0,
  errors: 0,
  latency: []
};

// В handler
metrics.requests++;
const start = Date.now();

try {
  const result = await operation();
  metrics.latency.push(Date.now() - start);
  return result;
} catch (error) {
  metrics.errors++;
  throw error;
}
```

### Алерты

```typescript
// При критических ошибках
if (error.code === 'DATABASE_ERROR') {
  // Отправить в Sentry с высоким приоритетом
  captureException(error, {
    level: 'fatal',
    tags: { alert: 'database_down' }
  });

  // Опционально: webhook в Telegram
  await sendAlert(`Database error: ${error.message}`);
}
```

## Ключевые файлы

| Файл | Назначение |
|------|------------|
| `logger-error-lib/src/logger.ts` | Winston логгер |
| `trace-lib/src/trace.ts` | Tracing utilities |
| `redis/src/helpers/sentryTracing.ts` | Sentry integration |
| `ydb/src/helpers/sqlErrorAnalyzer.ts` | SQL error analysis |

## Best Practices

1. **Структурированные логи** - JSON формат
2. **Correlation ID** во всех запросах
3. **Контекст в ошибках** - что, где, когда
4. **Уровни логирования** - debug/info/warn/error
5. **Не логируй sensitive data** - токены, пароли
6. **Sentry для продакшена** - алерты и трейсинг
7. **Health checks** для мониторинга
