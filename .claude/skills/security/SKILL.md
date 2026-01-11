---
name: security
description: Безопасность в @mrg0773. Шифрование, аутентификация, блокировки, секреты. Используй для криптографии, JWT, защиты данных.
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# Security

## Шифрование (@mrg0773/encryption-lib)

### AES шифрование

```typescript
import { EncryptionService } from '@mrg0773/encryption-lib';

const encryption = new EncryptionService(process.env.ENCRYPTION_KEY);

// Шифрование
const encrypted = await encryption.encrypt('sensitive data');
// { iv: '...', data: '...', tag: '...' }

// Расшифровка
const decrypted = await encryption.decrypt(encrypted);
// 'sensitive data'
```

### Хеширование

```typescript
import { HashService } from '@mrg0773/encryption-lib';

// Хеш пароля
const hash = await HashService.hashPassword('password123');

// Проверка
const isValid = await HashService.verifyPassword('password123', hash);
```

### HMAC

```typescript
import { HmacService } from '@mrg0773/encryption-lib';

// Создать подпись
const signature = HmacService.sign(data, secretKey);

// Проверить
const isValid = HmacService.verify(data, signature, secretKey);
```

## Распределённые блокировки (@mrg0773/lock-lib)

### Простая блокировка

```typescript
import { LockService } from '@mrg0773/lock-lib';

const lock = new LockService(redisClient);

// Получить блокировку
const result = await lock.acquire('resource:123', {
  ttl: 30000,  // 30 секунд
  retries: 3,
  retryDelay: 100
});

if (result.isOk()) {
  try {
    // Критическая секция
    await processResource('123');
  } finally {
    await lock.release('resource:123');
  }
} else {
  console.log('Could not acquire lock');
}
```

### With Lock (автоматическое освобождение)

```typescript
const result = await lock.withLock('resource:123', async () => {
  // Блокировка автоматически освободится
  return await processResource('123');
}, { ttl: 30000 });

if (result.isOk()) {
  console.log('Result:', result.value);
}
```

### Mutex для конкурентного доступа

```typescript
import { Mutex } from '@mrg0773/lock-lib';

const mutex = new Mutex(redisClient, 'my-mutex');

await mutex.runExclusive(async () => {
  // Только один процесс одновременно
  await updateSharedState();
});
```

## JWT Аутентификация

### Создание токена

```typescript
import jwt from 'jsonwebtoken';

function createToken(userId: string, roles: string[]): string {
  return jwt.sign(
    {
      sub: userId,
      roles,
      iat: Date.now()
    },
    process.env.JWT_SECRET,
    { expiresIn: '24h' }
  );
}
```

### Проверка токена

```typescript
import { ok, err, Result } from '@mrg0773/logger-error';

interface TokenPayload {
  sub: string;
  roles: string[];
  iat: number;
}

function verifyToken(token: string): Result<TokenPayload, AppError> {
  try {
    const payload = jwt.verify(token, process.env.JWT_SECRET) as TokenPayload;
    return ok(payload);
  } catch (error) {
    if (error.name === 'TokenExpiredError') {
      return err(createAuthError('Token expired'));
    }
    return err(createAuthError('Invalid token'));
  }
}
```

### Middleware

```typescript
async function authMiddleware(event: HttpEvent): Promise<Result<TokenPayload, AppError>> {
  const authHeader = event.headers['authorization'];

  if (!authHeader?.startsWith('Bearer ')) {
    return err(createAuthError('Missing authorization header'));
  }

  const token = authHeader.slice(7);
  return verifyToken(token);
}

// Использование
export async function handler(event: HttpEvent) {
  const authResult = await authMiddleware(event);

  if (authResult.isErr()) {
    return {
      statusCode: 401,
      body: JSON.stringify({ error: authResult.error.message })
    };
  }

  const user = authResult.value;
  // Продолжить с авторизованным пользователем
}
```

## Telegram Auth

```typescript
import crypto from 'crypto';

interface TelegramAuthData {
  id: number;
  first_name: string;
  auth_date: number;
  hash: string;
  [key: string]: unknown;
}

function verifyTelegramAuth(data: TelegramAuthData, botToken: string): boolean {
  const { hash, ...checkData } = data;

  // Создать строку для проверки
  const checkString = Object.keys(checkData)
    .sort()
    .map(key => `${key}=${checkData[key]}`)
    .join('\n');

  // Создать ключ из токена бота
  const secretKey = crypto
    .createHash('sha256')
    .update(botToken)
    .digest();

  // Вычислить HMAC
  const hmac = crypto
    .createHmac('sha256', secretKey)
    .update(checkString)
    .digest('hex');

  return hmac === hash;
}
```

## Секреты

### Environment Variables

```typescript
// Загрузка из .env (только локально!)
import 'dotenv/config';

// В коде
const apiKey = process.env.API_KEY;
if (!apiKey) {
  throw new Error('API_KEY is required');
}
```

### Yandex Lockbox

```typescript
import { LockboxClient } from '@yandex-cloud/nodejs-sdk';

async function getSecret(secretId: string, key: string): Promise<string> {
  const client = new LockboxClient();

  const payload = await client.get({
    secretId,
    versionId: 'latest'
  });

  const entry = payload.entries.find(e => e.key === key);
  return entry?.textValue || '';
}
```

## Валидация входных данных

```typescript
import { z } from 'zod';

// Схема
const UserInputSchema = z.object({
  email: z.string().email(),
  password: z.string().min(8).max(100),
  name: z.string().min(1).max(100)
});

// Валидация
function validateInput(data: unknown): Result<UserInput, AppError> {
  const result = UserInputSchema.safeParse(data);

  if (!result.success) {
    return err(createValidationError('Invalid input', {
      context: { errors: result.error.errors }
    }));
  }

  return ok(result.data);
}
```

## SQL Injection Protection

```typescript
// ❌ ПЛОХО - конкатенация
const query = `SELECT * FROM users WHERE id = '${userId}'`;

// ✅ ХОРОШО - параметризованный запрос
const result = await ydb.query`
  SELECT * FROM users WHERE id = ${userId}
`;

// YDB автоматически экранирует параметры
```

## Rate Limiting

```typescript
import { RateLimiter } from '@mrg0773/redis';

const limiter = new RateLimiter({
  tokensPerSecond: 10,
  bucketSize: 100
});

async function checkRateLimit(userId: string): Promise<Result<void, AppError>> {
  const allowed = await limiter.tryConsume(userId, 1);

  if (!allowed) {
    return err(createAppError('API_ERROR', 'Rate limit exceeded', {
      context: { retryAfter: 60 }
    }));
  }

  return ok(undefined);
}
```

## Ключевые файлы

| Файл | Назначение |
|------|------------|
| `encryption-lib/src/crypto.ts` | AES шифрование |
| `encryption-lib/src/hash.ts` | Хеширование |
| `lock-lib/src/lock.ts` | Распределённые блокировки |
| `lock-lib/src/mutex.ts` | Mutex |

## Security Checklist

- [ ] Секреты в Lockbox, не в коде
- [ ] HTTPS везде
- [ ] JWT с коротким временем жизни
- [ ] Rate limiting на API
- [ ] Валидация всех входных данных
- [ ] Параметризованные SQL запросы
- [ ] Шифрование sensitive данных
- [ ] Логирование auth событий
- [ ] Регулярная ротация ключей
