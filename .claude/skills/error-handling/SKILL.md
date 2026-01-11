---
name: error-handling
description: Обработка ошибок в @mrg0773 с neverthrow. Result типы, error factories, AppError. Используй для обработки ошибок, создания типизированных ошибок, работы с Result.
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# Error Handling с neverthrow

## Философия

**НЕ используем try-catch!** Используем `Result<T, E>` из neverthrow.

```typescript
// ❌ ПЛОХО - try-catch
async function getUser(id: string): Promise<User> {
  try {
    const user = await db.query(`SELECT * FROM users WHERE id = ${id}`);
    return user;
  } catch (error) {
    console.error(error);
    throw error;
  }
}

// ✅ ХОРОШО - neverthrow
async function getUser(id: string): Promise<Result<User, AppError>> {
  const result = await db.query`SELECT * FROM users WHERE id = ${id}`;

  if (result.isErr()) {
    return err(createDatabaseError('Failed to get user'));
  }

  return ok(result.value);
}
```

## Базовые операции

```typescript
import { ok, err, Result, ResultAsync } from '@mrg0773/logger-error';

// Успешный результат
const success = ok({ id: 1, name: 'John' });

// Ошибка
const failure = err(new Error('Something went wrong'));

// Проверка
if (success.isOk()) {
  console.log(success.value); // { id: 1, name: 'John' }
}

if (failure.isErr()) {
  console.log(failure.error); // Error: Something went wrong
}
```

## AppError

Стандартизированный тип ошибки:

```typescript
import { AppError, createAppError } from '@mrg0773/logger-error';

interface AppError {
  code: ErrorCode;
  message: string;
  context?: Record<string, unknown>;
  cause?: Error;
}

type ErrorCode =
  | 'UNKNOWN_ERROR'
  | 'VALIDATION_ERROR'
  | 'NETWORK_ERROR'
  | 'HTTP_ERROR'
  | 'PARSE_ERROR'
  | 'API_ERROR'
  | 'TIMEOUT_ERROR'
  | 'AUTH_ERROR'
  | 'NOT_FOUND_ERROR'
  | 'DATABASE_ERROR'
  | 'REDIS_ERROR'
  | 'CONFIG_ERROR';
```

## Error Factories

```typescript
import {
  createAppError,
  createValidationError,
  createNetworkError,
  createHttpError,
  createTimeoutError,
  createAuthError,
  createDatabaseError,
  createRedisError,
  createNotFoundError
} from '@mrg0773/logger-error';

// Общая ошибка
const error = createAppError('API_ERROR', 'External API failed', {
  context: { endpoint: '/api/users', status: 500 }
});

// Валидация
const validationError = createValidationError('Invalid email format', {
  context: { field: 'email', value: 'invalid' }
});

// База данных
const dbError = createDatabaseError('Query timeout', {
  context: { query: 'SELECT...', duration: 30000 }
});

// Не найдено
const notFound = createNotFoundError('User not found', {
  context: { userId: '123' }
});
```

## errWithLog

Создаёт ошибку И логирует её:

```typescript
import { errWithLog } from '@mrg0773/logger-error';

async function processOrder(orderId: string): Promise<Result<Order, AppError>> {
  const order = await db.getOrder(orderId);

  if (!order) {
    // Автоматически логируется!
    return errWithLog('NOT_FOUND_ERROR', 'Order not found', {
      context: { orderId }
    });
  }

  return ok(order);
}
```

## Цепочки операций

### map - трансформация успешного значения

```typescript
const result = ok(5)
  .map(x => x * 2)  // ok(10)
  .map(x => x + 1); // ok(11)
```

### mapErr - трансформация ошибки

```typescript
const result = err('error')
  .mapErr(e => new AppError(e)); // err(AppError)
```

### andThen - цепочка Result

```typescript
const result = await getUserId(token)
  .andThen(userId => getUser(userId))
  .andThen(user => getOrders(user.id));
```

### orElse - обработка ошибки

```typescript
const result = await getFromCache(key)
  .orElse(() => getFromDatabase(key)); // Fallback
```

## Обработка неизвестных ошибок

```typescript
import { handleUnknownError, isAppError } from '@mrg0773/logger-error';

try {
  await riskyExternalCall();
} catch (unknown) {
  const appError = handleUnknownError(unknown);
  // Теперь это типизированный AppError

  if (isAppError(appError)) {
    console.log(appError.code, appError.message);
  }
}
```

## Паттерны использования

### Repository Layer

```typescript
class UserRepository {
  async findById(id: string): Promise<Result<User | null, AppError>> {
    const result = await this.ydb.query`
      SELECT * FROM users WHERE id = ${id}
    `;

    if (result.isErr()) {
      return errWithLog('DATABASE_ERROR', 'Failed to find user', {
        context: { userId: id }
      });
    }

    return ok(result.value[0] || null);
  }

  async create(user: CreateUserDto): Promise<Result<User, AppError>> {
    // Валидация
    const validation = this.validate(user);
    if (validation.isErr()) {
      return validation;
    }

    // Сохранение
    const result = await this.ydb.upsert('users', user);
    if (result.isErr()) {
      return errWithLog('DATABASE_ERROR', 'Failed to create user');
    }

    return ok(user);
  }
}
```

### Service Layer

```typescript
class OrderService {
  async placeOrder(dto: PlaceOrderDto): Promise<Result<Order, AppError>> {
    // Получить пользователя
    const userResult = await this.userRepo.findById(dto.userId);
    if (userResult.isErr()) return userResult;

    const user = userResult.value;
    if (!user) {
      return errWithLog('NOT_FOUND_ERROR', 'User not found');
    }

    // Проверить баланс
    if (user.balance < dto.total) {
      return errWithLog('VALIDATION_ERROR', 'Insufficient balance', {
        context: { required: dto.total, available: user.balance }
      });
    }

    // Создать заказ
    return await this.orderRepo.create({
      userId: user.id,
      items: dto.items,
      total: dto.total
    });
  }
}
```

### Handler Layer (Serverless)

```typescript
export async function handler(event: APIGatewayEvent) {
  const result = await orderService.placeOrder(JSON.parse(event.body));

  if (result.isErr()) {
    const error = result.error;

    // Маппинг кодов ошибок на HTTP статусы
    const statusMap: Record<ErrorCode, number> = {
      VALIDATION_ERROR: 400,
      AUTH_ERROR: 401,
      NOT_FOUND_ERROR: 404,
      DATABASE_ERROR: 500,
      // ...
    };

    return {
      statusCode: statusMap[error.code] || 500,
      body: JSON.stringify({
        error: error.code,
        message: error.message
      })
    };
  }

  return {
    statusCode: 200,
    body: JSON.stringify(result.value)
  };
}
```

## Type Guards

```typescript
import { isAppError } from '@mrg0773/logger-error';

function handleError(error: unknown) {
  if (isAppError(error)) {
    // TypeScript знает что это AppError
    console.log(error.code, error.message, error.context);
  } else {
    // Неизвестная ошибка
    console.log('Unknown error:', error);
  }
}
```

## safeRepoCall Wrapper

Для оборачивания операций с БД:

```typescript
import { safeRepoCall } from '@mrg0773/tgferma';

async function getUser(id: string) {
  return safeRepoCall(
    () => userRepo.findById(id),
    'getUser',
    { userId: id }
  );
}
// Автоматически логирует ошибки с контекстом
```

## Best Practices

1. **Всегда возвращай Result** из функций которые могут упасть
2. **Используй error factories** для консистентности
3. **Добавляй context** с полезной информацией
4. **errWithLog** для важных ошибок (автолог)
5. **Не глотай ошибки** - пробрасывай вверх
6. **Type guards** для unknown ошибок
