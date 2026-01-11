---
name: redis-patterns
description: Паттерны работы с Redis в @mrg0773. Кэширование, Circuit Breaker, Rate Limiter, сессии. Используй для кэша, очередей, блокировок, rate limiting.
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# Redis Patterns

## Подключение

```typescript
import { RedisServiceDirect } from '@mrg0773/redis';

// Создание клиента
const redis = new RedisServiceDirect(10); // database index
await redis.connect();

// Базовые операции
const value = await redis.get('key');
await redis.set('key', 'value', 3600); // TTL в секундах
await redis.del('key');
```

## Стратегии кэширования

### Write-Through (синхронная запись)

```typescript
import { CacheService } from '@mrg0773/redis';

const cache = await CacheService.create(10, {
  namespace: 'myapp',
  strategy: 'write-through',
  ttl: 300
});

// При записи - кэш и БД обновляются одновременно
await cache.set('user:123', userData);
```

### Write-Behind (отложенная запись)

```typescript
const cache = await CacheService.create(10, {
  strategy: 'write-behind',
  ttl: 300
});

// Кэш обновляется сразу, БД - асинхронно
await cache.set('user:123', userData);
```

### Write-Around (пишем мимо кэша)

```typescript
const cache = await CacheService.create(10, {
  strategy: 'write-around'
});

// Запись идёт напрямую в БД, кэш не обновляется
// Используй когда данные редко читаются после записи
```

## Circuit Breaker

Защита от каскадных сбоев:

```typescript
import { CircuitBreaker } from '@mrg0773/redis';

const breaker = new CircuitBreaker({
  threshold: 5,      // Сколько ошибок для открытия
  timeout: 30000,    // Сколько ждать перед проверкой
  halfOpenRequests: 3 // Запросов в half-open
});

// Состояния:
// CLOSED - нормальная работа
// OPEN - отказ (fail fast)
// HALF_OPEN - проверка восстановления

const result = await breaker.execute(async () => {
  return await riskyOperation();
});

// Проверка состояния
if (breaker.isOpen()) {
  console.log('Service unavailable');
}
```

## Rate Limiter

Token bucket алгоритм:

```typescript
import { RateLimiter } from '@mrg0773/redis';

const limiter = new RateLimiter({
  tokensPerSecond: 10,  // Скорость пополнения
  bucketSize: 100       // Максимум токенов
});

// Проверка и потребление
const canProceed = await limiter.tryConsume(1);
if (!canProceed) {
  throw new Error('Rate limit exceeded');
}

// Для API endpoints
async function handleRequest(req, res) {
  const allowed = await limiter.tryConsume(1);
  if (!allowed) {
    return res.status(429).json({ error: 'Too many requests' });
  }
  // Process request
}
```

## Concurrency Control

Ограничение параллельных операций:

```typescript
import { ConcurrencyControl } from '@mrg0773/redis';

const control = new ConcurrencyControl({
  maxConcurrent: 5,
  queueTimeout: 10000
});

// Только 5 параллельных операций
await control.run(async () => {
  await heavyOperation();
});
```

## Паттерны использования

### Cache-Aside (Lazy Loading)

```typescript
async function getUser(userId: string) {
  // 1. Проверить кэш
  const cached = await redis.get(`user:${userId}`);
  if (cached) {
    return JSON.parse(cached);
  }

  // 2. Загрузить из БД
  const user = await db.getUser(userId);

  // 3. Сохранить в кэш
  await redis.set(`user:${userId}`, JSON.stringify(user), 3600);

  return user;
}
```

### Distributed Lock

```typescript
import { LockService } from '@mrg0773/lock-lib';

const lock = new LockService(redis);

// Получить блокировку
const acquired = await lock.acquire('resource:123', {
  ttl: 30000,  // 30 секунд
  retries: 3
});

if (acquired.isOk()) {
  try {
    // Критическая секция
    await criticalOperation();
  } finally {
    await lock.release('resource:123');
  }
}
```

### Session Storage

```typescript
// Сохранить сессию
await redis.set(`session:${sessionId}`, JSON.stringify({
  userId: 123,
  roles: ['user'],
  expiresAt: Date.now() + 3600000
}), 3600);

// Получить сессию
const session = await redis.get(`session:${sessionId}`);
if (session) {
  return JSON.parse(session);
}
```

### Pub/Sub

```typescript
// Подписка
redis.subscribe('notifications', (message) => {
  console.log('Received:', message);
});

// Публикация
await redis.publish('notifications', JSON.stringify({
  type: 'user_created',
  userId: 123
}));
```

## Обработка ошибок

```typescript
import { ok, err, Result } from '@mrg0773/logger-error';

async function getCached<T>(key: string): Promise<Result<T | null, RedisError>> {
  const result = await redis.get(key);

  if (result.isErr()) {
    return err(createRedisError('Cache read failed', {
      context: { key }
    }));
  }

  if (!result.value) {
    return ok(null);
  }

  return ok(JSON.parse(result.value));
}
```

## Sentry Integration

```typescript
import { createSentrySpan } from '@mrg0773/redis';

// Автоматический трейсинг операций
const span = createSentrySpan('redis.get', 'user:123');
const result = await redis.get('user:123');
span.setStatus(result.isOk() ? 'ok' : 'error');
span.finish();
```

## Ключевые файлы

| Файл | Назначение |
|------|------------|
| `redis/src/services/redisServiceDirect.ts` | Прямой ioredis клиент |
| `redis/src/services/cacheService.ts` | Стратегии кэширования |
| `redis/src/services/circuitBreaker.ts` | Circuit Breaker |
| `redis/src/services/rateLimiter.ts` | Rate Limiting |
| `redis/src/helpers/sentryTracing.ts` | Sentry интеграция |

## Best Practices

1. **Используй namespaces** для ключей: `myapp:users:123`
2. **Всегда ставь TTL** - избегай memory leaks
3. **Circuit Breaker** для внешних сервисов
4. **Rate Limiter** на API endpoints
5. **Graceful degradation** - приложение работает без Redis
6. **Мониторь память** - Redis in-memory database
